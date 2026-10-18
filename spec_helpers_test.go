package voiceml_test

// Helpers that load the vendored public OpenAPI document
// (testdata/voiceml-openapi.json) for the spec-driven contract tests.

import (
	"encoding/json"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
)

const specFixture = "testdata/voiceml-openapi.json"

// Host labels used by the contract tests.
const (
	hostDefault       = "default"
	hostMessaging     = "messaging"
	hostConversations = "conversations"
)

type specParam struct {
	Required bool
	Array    bool
}

type specBody struct {
	Required map[string]bool
	Props    map[string]specParam
}

type specOp struct {
	Method    string
	Path      string
	ID        string
	Tags      []string
	NoAuth    bool
	Hosts     map[string]bool
	Query     map[string]specParam
	Bodies    map[string]specBody // keyed by content type
	Success   map[string]any      // 2xx status -> resolved JSON schema (nil when the body is not JSON)
	re        *regexp.Regexp
	literalSz int
}

type specDoc struct {
	Version string
	Ops     []*specOp
	root    map[string]any
}

var (
	specOnce   sync.Once
	specLoaded *specDoc
	specErr    error
)

func loadSpec(t *testing.T) *specDoc {
	t.Helper()
	specOnce.Do(func() {
		var raw []byte
		raw, specErr = os.ReadFile(specFixture)
		if specErr != nil {
			return
		}
		var root map[string]any
		if specErr = json.Unmarshal(raw, &root); specErr != nil {
			return
		}
		specLoaded = buildSpec(root)
	})
	if specErr != nil {
		t.Fatalf("load %s: %v", specFixture, specErr)
	}
	return specLoaded
}

func buildSpec(root map[string]any) *specDoc {
	d := &specDoc{root: root}
	if info, ok := root["info"].(map[string]any); ok {
		d.Version, _ = info["version"].(string)
	}
	paths, _ := root["paths"].(map[string]any)
	keys := make([]string, 0, len(paths))
	for k := range paths {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, p := range keys {
		item, _ := paths[p].(map[string]any)
		var shared []any
		if ps, ok := item["parameters"].([]any); ok {
			shared = ps
		}
		for _, m := range []string{"get", "post", "put", "patch", "delete"} {
			raw, ok := item[m].(map[string]any)
			if !ok {
				continue
			}
			op := &specOp{
				Method:  strings.ToUpper(m),
				Path:    p,
				Query:   map[string]specParam{},
				Bodies:  map[string]specBody{},
				Success: map[string]any{},
				Hosts:   map[string]bool{},
			}
			op.ID, _ = raw["operationId"].(string)
			for _, tg := range raw["tags"].([]any) {
				op.Tags = append(op.Tags, tg.(string))
			}
			if sec, ok := raw["security"].([]any); ok && len(sec) == 0 {
				op.NoAuth = true
			}
			for _, tg := range op.Tags {
				switch {
				case strings.HasPrefix(tg, "ConversationsV1"):
					op.Hosts[hostConversations] = true
				case tg == "MessagingV1Service":
					op.Hosts[hostMessaging] = true
				default:
					op.Hosts[hostDefault] = true
				}
			}
			params := append(append([]any{}, shared...), asSlice(raw["parameters"])...)
			for _, prm := range params {
				pm := d.resolve(prm)
				if pm["in"] != "query" {
					continue
				}
				name := pm["name"].(string)
				req, _ := pm["required"].(bool)
				op.Query[name] = specParam{Required: req, Array: d.isArray(pm["schema"])}
			}
			if rb, ok := raw["requestBody"].(map[string]any); ok {
				rb = d.resolve(rb)
				content, _ := rb["content"].(map[string]any)
				for ct, media := range content {
					sch := d.merged(asMap(media)["schema"])
					body := specBody{Required: map[string]bool{}, Props: map[string]specParam{}}
					for _, r := range asSlice(sch["required"]) {
						body.Required[r.(string)] = true
					}
					for name, ps := range asMap(sch["properties"]) {
						body.Props[name] = specParam{Required: body.Required[name], Array: d.isArray(ps)}
					}
					op.Bodies[ct] = body
				}
			}
			for code, r := range asMap(raw["responses"]) {
				if !strings.HasPrefix(code, "2") {
					continue
				}
				resp := d.resolve(r)
				content := asMap(resp["content"])
				if media, ok := content["application/json"]; ok {
					op.Success[code] = d.merged(asMap(media)["schema"])
				} else {
					op.Success[code] = nil
				}
			}
			op.re, op.literalSz = templateRegexp(p)
			d.Ops = append(d.Ops, op)
		}
	}
	return d
}

func templateRegexp(p string) (*regexp.Regexp, int) {
	var b strings.Builder
	b.WriteString("^")
	literal := 0
	for len(p) > 0 {
		i := strings.IndexByte(p, '{')
		if i < 0 {
			b.WriteString(regexp.QuoteMeta(p))
			literal += len(p)
			break
		}
		b.WriteString(regexp.QuoteMeta(p[:i]))
		literal += i
		j := strings.IndexByte(p[i:], '}')
		b.WriteString(`[^/]+`)
		p = p[i+j+1:]
	}
	b.WriteString("$")
	return regexp.MustCompile(b.String()), literal
}

// match returns the most specific operation for (method, escaped path).
func (d *specDoc) match(method, path string) *specOp {
	var best *specOp
	for _, op := range d.Ops {
		if op.Method != method || !op.re.MatchString(path) {
			continue
		}
		if best == nil || op.literalSz > best.literalSz {
			best = op
		}
	}
	return best
}

func (d *specDoc) resolve(v any) map[string]any {
	m := asMap(v)
	for i := 0; i < 32; i++ {
		ref, ok := m["$ref"].(string)
		if !ok {
			return m
		}
		var cur any = d.root
		for _, part := range strings.Split(strings.TrimPrefix(ref, "#/"), "/") {
			cur = asMap(cur)[part]
		}
		target := asMap(cur)
		if len(m) > 1 {
			merged := map[string]any{}
			for k, v := range target {
				merged[k] = v
			}
			for k, v := range m {
				if k != "$ref" {
					merged[k] = v
				}
			}
			target = merged
		}
		m = target
	}
	return m
}

// merged resolves $ref and flattens allOf into a single object schema.
func (d *specDoc) merged(v any) map[string]any {
	m := d.resolve(v)
	all, ok := m["allOf"].([]any)
	if !ok {
		return m
	}
	out := map[string]any{}
	for k, v := range m {
		if k != "allOf" {
			out[k] = v
		}
	}
	props := map[string]any{}
	for k, v := range asMap(m["properties"]) {
		props[k] = v
	}
	var req []any
	req = append(req, asSlice(m["required"])...)
	for _, sub := range all {
		sm := d.merged(sub)
		for k, v := range asMap(sm["properties"]) {
			props[k] = v
		}
		req = append(req, asSlice(sm["required"])...)
		for _, k := range []string{"type", "items", "additionalProperties"} {
			if _, have := out[k]; !have {
				if v, ok := sm[k]; ok {
					out[k] = v
				}
			}
		}
	}
	out["properties"] = props
	out["required"] = req
	return out
}

// types returns the JSON types a schema admits (without "null").
func (d *specDoc) types(v any) (types map[string]bool, nullable bool) {
	m := d.merged(v)
	types = map[string]bool{}
	switch t := m["type"].(type) {
	case string:
		types[t] = true
	case []any:
		for _, x := range t {
			types[x.(string)] = true
		}
	}
	for _, k := range []string{"oneOf", "anyOf"} {
		for _, sub := range asSlice(m[k]) {
			st, sn := d.types(sub)
			for x := range st {
				types[x] = true
			}
			nullable = nullable || sn
		}
	}
	if n, _ := m["nullable"].(bool); n {
		nullable = true
	}
	if types["null"] {
		nullable = true
		delete(types, "null")
	}
	if len(types) == 0 && m["properties"] != nil {
		types["object"] = true
	}
	return types, nullable
}

func (d *specDoc) isArray(v any) bool {
	t, _ := d.types(v)
	return t["array"]
}

func asMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	if m == nil {
		return map[string]any{}
	}
	return m
}

func asSlice(v any) []any {
	s, _ := v.([]any)
	return s
}
