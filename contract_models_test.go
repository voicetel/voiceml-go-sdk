package voiceml_test

// Response-model half of the spec contract: every SDK method's return type
// is compared field-by-field with the 2xx JSON schema of the operation it
// calls, and spec-derived sample bodies (all-null, and all-present with
// unknown enum values and extra fields) must decode into it.

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// extraModelFields lists Go response fields the 2xx schema does not declare,
// each with the reason it is kept. Keyed by "<operationId> <json path>".
var extraModelFields = map[string]string{
	"health.check .failures": "the 503 body carries failures; the same struct models both",
}

// documentedSubsetStructs are Go structs that model a spec `type: object`
// with no declared properties. Each names the keys the API documents in
// prose; it is also the request shape, so it stays a struct.
var documentedSubsetStructs = map[string]string{
	"voiceml.AssistantsV1CustomerAI": "customer_ai documents perception_engine_enabled and personalization_engine_enabled; the same struct is the request shape",
	"voiceml.VoiceV1Meta":            "the shared /v1 meta envelope (components/schemas/VoiceV1Meta); the sender-pool list leaves it untyped",
}

var rawMessageType = reflect.TypeOf(json.RawMessage(nil))

type modelChecker struct {
	spec   *specDoc
	issues []string
}

func (mc *modelChecker) fail(format string, args ...any) {
	mc.issues = append(mc.issues, fmt.Sprintf(format, args...))
}

// jsonFields flattens a struct's JSON-visible fields (embedded structs
// without a tag are promoted, as encoding/json does).
func jsonFields(t reflect.Type) map[string]reflect.Type {
	out := map[string]reflect.Type{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("json")
		name, _, _ := strings.Cut(tag, ",")
		if f.Anonymous && tag == "" {
			et := f.Type
			if et.Kind() == reflect.Ptr {
				et = et.Elem()
			}
			for k, v := range jsonFields(et) {
				if _, ok := out[k]; !ok {
					out[k] = v
				}
			}
			continue
		}
		if !f.IsExported() || name == "-" {
			continue
		}
		if name == "" {
			name = f.Name
		}
		out[name] = f.Type
	}
	return out
}

// alternatives returns the oneOf/anyOf branches of a schema, if any.
func (mc *modelChecker) alternatives(schema any) []any {
	sch := mc.spec.merged(schema)
	if alts := asSlice(sch["oneOf"]); len(alts) > 0 {
		return alts
	}
	return asSlice(sch["anyOf"])
}

// bestAlternative picks the oneOf branch the Go type models most closely.
func (mc *modelChecker) bestAlternative(opID string, gt reflect.Type, schema any, path string) (any, []string) {
	var best any
	var bestIssues []string
	for i, alt := range mc.alternatives(schema) {
		sub := &modelChecker{spec: mc.spec}
		sub.check(opID, gt, alt, path)
		if i == 0 || len(sub.issues) < len(bestIssues) {
			best, bestIssues = alt, sub.issues
		}
	}
	return best, bestIssues
}

func (mc *modelChecker) check(opID string, gt reflect.Type, schema any, path string) {
	for gt.Kind() == reflect.Ptr {
		gt = gt.Elem()
	}
	if gt == rawMessageType || (gt.Kind() == reflect.Interface && gt.NumMethod() == 0) {
		return
	}
	if len(mc.alternatives(schema)) > 0 {
		_, issues := mc.bestAlternative(opID, gt, schema, path)
		mc.issues = append(mc.issues, issues...)
		return
	}
	sch := mc.spec.merged(schema)
	types, _ := mc.spec.types(sch)
	if len(types) == 0 {
		return // unconstrained schema
	}
	if len(types) > 1 {
		mc.fail("%s %s: spec allows %v; Go type %s cannot hold all of them", opID, path, keysOf(types), gt)
		return
	}
	var want string
	for k := range types {
		want = k
	}
	switch want {
	case "string":
		if gt.Kind() != reflect.String {
			mc.fail("%s %s: spec string, Go %s", opID, path, gt)
		}
	case "integer":
		switch gt.Kind() {
		case reflect.Int, reflect.Int32, reflect.Int64:
		default:
			mc.fail("%s %s: spec integer, Go %s", opID, path, gt)
		}
	case "number":
		if gt.Kind() != reflect.Float64 && gt.Kind() != reflect.Float32 {
			mc.fail("%s %s: spec number, Go %s", opID, path, gt)
		}
	case "boolean":
		if gt.Kind() != reflect.Bool {
			mc.fail("%s %s: spec boolean, Go %s", opID, path, gt)
		}
	case "array":
		if gt.Kind() != reflect.Slice {
			mc.fail("%s %s: spec array, Go %s", opID, path, gt)
			return
		}
		if items, ok := sch["items"]; ok {
			mc.check(opID, gt.Elem(), items, path+"[]")
		}
	case "object":
		props := asMap(sch["properties"])
		ap, hasAP := sch["additionalProperties"]
		switch gt.Kind() {
		case reflect.Map:
			if apSchema, ok := ap.(map[string]any); ok && hasAP {
				mc.check(opID, gt.Elem(), apSchema, path+"{}")
				return
			}
			if len(props) > 0 {
				for name, ps := range props {
					mc.check(opID, gt.Elem(), ps, path+"."+name)
				}
				return
			}
			// A bare object can carry any JSON value.
			if et := gt.Elem(); !(et.Kind() == reflect.Interface && et.NumMethod() == 0) && et != rawMessageType {
				mc.fail("%s %s: spec bare object, Go %s cannot decode arbitrary values", opID, path, gt)
			}
		case reflect.Struct:
			fields := jsonFields(gt)
			for name, ps := range props {
				ft, ok := fields[name]
				if !ok {
					mc.fail("%s %s.%s: declared by the spec, missing from %s", opID, path, name, gt)
					continue
				}
				mc.check(opID, ft, ps, path+"."+name)
			}
			if len(props) == 0 {
				if _, ok := documentedSubsetStructs[gt.String()]; !ok {
					mc.fail("%s %s: spec bare object, Go struct %s drops unknown keys", opID, path, gt)
				}
				return
			}
			for name := range fields {
				if _, ok := props[name]; ok {
					continue
				}
				key := opID + " " + path + "." + name
				if _, ok := extraModelFields[key]; ok {
					continue
				}
				mc.fail("%s %s.%s: %s field not declared by the spec", opID, path, name, gt)
			}
		default:
			mc.fail("%s %s: spec object, Go %s", opID, path, gt)
		}
	}
}

// sample builds a JSON value for schema. When nulls is true every nullable
// value is null; otherwise every value is present, enums carry an unknown
// member and every object gains an undeclared key.
func (mc *modelChecker) sample(gt reflect.Type, schema any, nulls bool, depth int) any {
	gt = derefType(gt)
	if len(mc.alternatives(schema)) > 0 {
		alt, _ := mc.bestAlternative("", gt, schema, "")
		return mc.sample(gt, alt, nulls, depth)
	}
	sch := mc.spec.merged(schema)
	types, nullable := mc.spec.types(sch)
	if nulls && nullable {
		return nil
	}
	if depth > 8 {
		return nil
	}
	if _, ok := sch["enum"]; ok && types["string"] && !nulls {
		return "value-added-after-this-sdk-release"
	}
	switch {
	case types["object"]:
		obj := map[string]any{}
		props := asMap(sch["properties"])
		var fields map[string]reflect.Type
		if gt.Kind() == reflect.Struct {
			fields = jsonFields(gt)
		}
		for name, ps := range props {
			ft := reflect.TypeOf((*any)(nil)).Elem()
			if f, ok := fields[name]; ok {
				ft = f
			} else if gt.Kind() == reflect.Map {
				ft = gt.Elem()
			}
			obj[name] = mc.sample(ft, ps, nulls, depth+1)
		}
		ap, hasAP := sch["additionalProperties"].(map[string]any)
		switch {
		case hasAP:
			et := reflect.TypeOf((*any)(nil)).Elem()
			if gt.Kind() == reflect.Map {
				et = gt.Elem()
			}
			obj["dynamic_key"] = mc.sample(et, ap, nulls, depth+1)
		case len(props) == 0:
			if _, ok := documentedSubsetStructs[gt.String()]; !ok {
				obj["number"] = 3
				obj["flag"] = true
				obj["nested"] = map[string]any{"list": []any{1, "two"}}
			}
		case !nulls && gt.Kind() == reflect.Map:
			// A keyed map of declared string properties (subresource_uris,
			// links) gains a new key of the same kind.
			obj["key_added_after_this_sdk_release"] = "text"
		case !nulls:
			obj["field_added_after_this_sdk_release"] = map[string]any{"x": 1}
		}
		return obj
	case types["array"]:
		et := reflect.TypeOf((*any)(nil)).Elem()
		if gt.Kind() == reflect.Slice {
			et = gt.Elem()
		}
		return []any{mc.sample(et, sch["items"], nulls, depth+1)}
	case types["integer"]:
		return 7
	case types["number"]:
		return 1.25
	case types["boolean"]:
		return false
	case types["string"]:
		return "text"
	}
	return nil
}

func TestSpecContractResponseModels(t *testing.T) {
	spec := loadSpec(t)
	calls := runContractWalk(t)
	mc := &modelChecker{spec: spec}
	seen := map[string]bool{}
	for _, call := range calls {
		if call.Ret == nil {
			continue
		}
		isModel := call.Ret.Kind() == reflect.Map ||
			(call.Ret.Kind() == reflect.Ptr && call.Ret.Elem().Kind() == reflect.Struct)
		if !isModel {
			continue
		}
		if strings.HasPrefix(call.Ret.String(), "*voiceml.Pager[") || call.Ret.String() == "*voiceml.Download" {
			continue
		}
		for _, r := range call.Reqs {
			op := spec.match(r.Method, r.Path)
			if op == nil {
				continue
			}
			key := op.ID + "|" + call.Ret.String()
			if seen[key] {
				continue
			}
			seen[key] = true
			codes := keysOf(op.Success)
			for _, code := range codes {
				schema := op.Success[code]
				if schema == nil {
					continue
				}
				mc.check(op.ID, call.Ret, schema, "")
				for _, nulls := range []bool{true, false} {
					body, err := json.Marshal(mc.sample(call.Ret, schema, nulls, 0))
					if err != nil {
						t.Fatalf("%s: marshal sample: %v", op.ID, err)
					}
					target := reflect.New(derefType(call.Ret))
					if err := json.Unmarshal(body, target.Interface()); err != nil {
						mc.fail("%s (%s %s.%s): decoding spec sample (nulls=%v) failed: %v", op.ID, code, call.Service, call.Method, nulls, err)
					}
				}
			}
		}
	}
	sort.Strings(mc.issues)
	issues := dedupe(mc.issues)
	if len(issues) > 0 {
		t.Errorf("%d response-model mismatches:\n  %s", len(issues), strings.Join(issues, "\n  "))
	}
}

func derefType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func dedupe(in []string) []string {
	var out []string
	for i, s := range in {
		if i == 0 || s != in[i-1] {
			out = append(out, s)
		}
	}
	return out
}
