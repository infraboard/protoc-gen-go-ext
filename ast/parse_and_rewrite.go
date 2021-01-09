// Copyright 2020 The searKing Author. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ast

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"google.golang.org/protobuf/compiler/protogen"

	pb "github.com/infraboard/protoc-gen-go-ext/extension/tag"
	"github.com/searKing/golang/go/reflect"
	strings_ "github.com/searKing/golang/go/strings"
)

type FieldInfo struct {
	FieldNameInProto string
	FieldNameInGo    string
	FieldTag         reflect.StructTag
	UpdateStrategy   pb.FieldTag_UpdateStrategy
}
type StructInfo struct {
	StructNameInProto string
	StructNameInGo    string
	FieldInfos        []FieldInfo
}

type FileInfo struct {
	FileName    string
	StructInfos []StructInfo
}

func (si *StructInfo) FindField(name string) (FieldInfo, bool) {
	for _, f := range si.FieldInfos {
		if f.FieldNameInGo == name {
			return f, true
		}
	}
	return FieldInfo{}, false
}

func WalkDescriptorProto(g *protogen.Plugin, dp *descriptor.DescriptorProto, typeNames []string) []StructInfo {
	var ss []StructInfo

	s := StructInfo{}
	s.StructNameInProto = dp.GetName()
	s.StructNameInGo = CamelCaseSlice(append(typeNames, CamelCase(dp.GetName())))

	//typeNames := []string{s.StructNameInGo}
	for _, field := range dp.GetField() {
		if field.GetOptions() == nil {
			continue
		}

		v, err := proto.GetExtension(field.Options, pb.E_FieldTag)
		if err != nil {
			continue
		}
		switch v := v.(type) {
		case *pb.FieldTag:
			tag := v.GetStructTag()
			tags, err := reflect.ParseStructTag(tag)
			if err != nil {
				g.Error(fmt.Errorf("failed to parse struct tag in field extension: %w", err))
				// ignore this tag
				continue
			}

			s.FieldInfos = append(s.FieldInfos, FieldInfo{
				FieldNameInProto: field.GetName(),
				FieldNameInGo:    CamelCase(field.GetName()),
				FieldTag:         *tags,
				UpdateStrategy:   v.GetUpdateStrategy(),
			})
		}
	}
	if len(s.FieldInfos) > 0 {
		ss = append(ss, s)
	}

	typeNames = append(typeNames, CamelCase(dp.GetName()))
	for _, nest := range dp.GetNestedType() {
		nestSs := WalkDescriptorProto(g, nest, typeNames)
		if len(nestSs) > 0 {
			ss = append(ss, nestSs...)
		}
	}
	return ss
}

func Rewrite(g *protogen.Plugin) {
	var protoFiles []FileInfo

	for _, protoFile := range g.Request.GetProtoFile() {
		if !strings_.SliceContains(g.Request.GetFileToGenerate(), protoFile.GetName()) {
			continue
		}
		f := FileInfo{}
		f.FileName = protoFile.GetName()

		for _, messageType := range protoFile.GetMessageType() {
			ss := WalkDescriptorProto(g, messageType, nil)
			if len(ss) > 0 {
				f.StructInfos = append(f.StructInfos, ss...)
			}
		}
		if len(f.StructInfos) > 0 {
			protoFiles = append(protoFiles, f)
		}
	}
	if len(protoFiles) == 0 {
		return
	}
	// g.Response() will generate files, so skip this step
	//if len(g.Response().GetFile()) == 0 {
	//	return
	//}

	rewriter := NewGenerator(protoFiles, g)
	for _, f := range g.Response().GetFile() {
		rewriter.ParseGoContent(f)
	}
	rewriter.Generate()
}
