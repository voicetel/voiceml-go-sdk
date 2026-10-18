package voiceml_test

// Spec-driven contract test. It loads the vendored public OpenAPI document
// (testdata/voiceml-openapi.json), reflectively calls every exported method
// on every service reachable from *voiceml.Client against three mock hosts
// (default, messaging, conversations), and asserts for every request:
//
//   - (method, escaped path incl. ".json") matches a spec operation;
//   - the request went to the host the operation's tag routes to;
//   - Authorization is present or absent exactly per the op's `security`;
//   - query, form and JSON keys are declared by the op (required ones
//     present, repeated keys only for array-typed params).
//
// It then asserts that every spec operation was reached and that every
// declared parameter can be sent, and decodes a spec-derived sample body
// (nulls, unknown enum values, extra fields) into each method's return type.

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	voiceml "github.com/voicetel/voiceml-go-sdk"
)

const (
	contractAccountSid = "ACXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
	contractAPIKey     = "contract-test-key"
)

// uncoverableOps lists spec operations the SDK cannot expose. It must stay
// empty unless an operation is impossible to model; every entry needs a reason.
var uncoverableOps = map[string]string{}

type wireRequest struct {
	Host        string
	Method      string
	Path        string
	RawQuery    string
	ContentType string
	Body        []byte
	Auth        string
}

type contractCall struct {
	Service string
	Method  string
	Ret     reflect.Type
	Reqs    []wireRequest
	Err     error
}

type wireRecorder struct {
	mu   sync.Mutex
	reqs []wireRequest
}

func (r *wireRecorder) handler(host string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, _ := io.ReadAll(req.Body)
		path, rawQuery, _ := strings.Cut(req.RequestURI, "?")
		r.mu.Lock()
		r.reqs = append(r.reqs, wireRequest{
			Host: host, Method: req.Method, Path: path, RawQuery: rawQuery,
			ContentType: req.Header.Get("Content-Type"), Body: body,
			Auth: req.Header.Get("Authorization"),
		})
		r.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
	})
}

func (r *wireRecorder) take() []wireRequest {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := r.reqs
	r.reqs = nil
	return out
}

var (
	contractOnce  sync.Once
	contractCalls []contractCall
)

// runContractWalk calls every public service method twice (bools true, then
// bools false) and returns what reached the wire.
func runContractWalk(t *testing.T) []contractCall {
	t.Helper()
	contractOnce.Do(func() {
		rec := &wireRecorder{}
		def := httptest.NewServer(rec.handler(hostDefault))
		msg := httptest.NewServer(rec.handler(hostMessaging))
		conv := httptest.NewServer(rec.handler(hostConversations))
		defer def.Close()
		defer msg.Close()
		defer conv.Close()
		zero := 0
		c, err := voiceml.NewClient(voiceml.ClientOptions{
			AccountSid: contractAccountSid, APIKey: contractAPIKey,
			BaseURL: def.URL, MessagingBaseURL: msg.URL, ConversationsBaseURL: conv.URL,
			MaxRetries: &zero,
		})
		if err != nil {
			t.Fatalf("NewClient: %v", err)
		}
		for _, boolVal := range []bool{true, false} {
			contractCalls = append(contractCalls, walkServices(reflect.ValueOf(c).Elem(), "Client", boolVal, rec)...)
		}
	})
	return contractCalls
}

func walkServices(v reflect.Value, path string, boolVal bool, rec *wireRecorder) []contractCall {
	var out []contractCall
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() || f.Type.Kind() != reflect.Ptr || f.Type.Elem().Kind() != reflect.Struct {
			continue
		}
		fv := v.Field(i)
		if fv.IsNil() {
			continue
		}
		svc := path + "." + f.Name
		out = append(out, callAllMethods(fv, svc, boolVal, rec)...)
		out = append(out, walkServices(fv.Elem(), svc, boolVal, rec)...)
	}
	return out
}

var ctxType = reflect.TypeOf((*context.Context)(nil)).Elem()

func callAllMethods(svc reflect.Value, name string, boolVal bool, rec *wireRecorder) []contractCall {
	var out []contractCall
	st := svc.Type()
	for i := 0; i < st.NumMethod(); i++ {
		m := st.Method(i)
		mt := m.Type
		args := []reflect.Value{svc}
		strN := 0
		for a := 1; a < mt.NumIn(); a++ {
			at := mt.In(a)
			if at == ctxType {
				args = append(args, reflect.ValueOf(context.Background()))
				continue
			}
			av := reflect.New(at).Elem()
			if at.Kind() == reflect.String {
				strN++
				av.SetString(fmt.Sprintf("P%d", strN))
			} else {
				fillValue(av, "A", boolVal, 0)
			}
			args = append(args, av)
		}
		rec.take()
		res := m.Func.Call(args)
		call := contractCall{Service: name, Method: m.Name}
		if mt.NumOut() > 0 {
			call.Ret = mt.Out(0)
		}
		for _, r := range res {
			if r.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) && !r.IsNil() {
				call.Err = r.Interface().(error)
			}
		}
		if len(res) > 0 {
			drainResult(res[0])
		}
		call.Reqs = rec.take()
		out = append(out, call)
	}
	return out
}

// drainResult advances pagers (so their first request is sent) and closes
// streamed bodies.
func drainResult(r reflect.Value) {
	if !r.IsValid() || (r.Kind() == reflect.Ptr && r.IsNil()) {
		return
	}
	if next := r.MethodByName("Next"); next.IsValid() && next.Type().NumIn() == 0 && next.Type().NumOut() == 1 && next.Type().Out(0).Kind() == reflect.Bool {
		for next.Call(nil)[0].Bool() {
		}
		return
	}
	if r.Kind() == reflect.Ptr && r.Elem().Kind() == reflect.Struct {
		if b := r.Elem().FieldByName("Body"); b.IsValid() && b.Type() == reflect.TypeOf((*io.ReadCloser)(nil)).Elem() && !b.IsNil() {
			_ = b.Interface().(io.ReadCloser).Close()
		}
	}
}

func fillValue(v reflect.Value, name string, boolVal bool, depth int) {
	if depth > 5 {
		return
	}
	switch v.Kind() {
	case reflect.String:
		v.SetString("S_" + name)
	case reflect.Int, reflect.Int64, reflect.Int32:
		v.SetInt(1)
	case reflect.Float64, reflect.Float32:
		v.SetFloat(0.5)
	case reflect.Bool:
		v.SetBool(boolVal)
	case reflect.Ptr:
		n := reflect.New(v.Type().Elem())
		fillValue(n.Elem(), name, boolVal, depth+1)
		v.Set(n)
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			v.SetBytes([]byte(`{"k":"v"}`))
			return
		}
		s := reflect.MakeSlice(v.Type(), 2, 2)
		for i := 0; i < 2; i++ {
			fillValue(s.Index(i), fmt.Sprintf("%s%d", name, i), boolVal, depth+1)
		}
		v.Set(s)
	case reflect.Map:
		m := reflect.MakeMap(v.Type())
		k := reflect.New(v.Type().Key()).Elem()
		fillValue(k, name+"K", boolVal, depth+1)
		val := reflect.New(v.Type().Elem()).Elem()
		fillValue(val, name+"V", boolVal, depth+1)
		m.SetMapIndex(k, val)
		v.Set(m)
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			if t.Field(i).IsExported() {
				fillValue(v.Field(i), t.Field(i).Name, boolVal, depth+1)
			}
		}
	case reflect.Interface:
		if v.NumMethod() == 0 {
			v.Set(reflect.ValueOf(map[string]any{"k": "v"}))
		}
	}
}

func expectedHost(service string) string {
	switch {
	case strings.HasPrefix(service, "Client.ConversationsV1"):
		return hostConversations
	case strings.HasPrefix(service, "Client.MessagingV1"):
		return hostMessaging
	}
	return hostDefault
}

func requestKeys(r wireRequest) (query url.Values, body map[string]int, err error) {
	query, err = url.ParseQuery(r.RawQuery)
	if err != nil {
		return nil, nil, err
	}
	body = map[string]int{}
	switch {
	case len(r.Body) == 0:
	case strings.HasPrefix(r.ContentType, "application/x-www-form-urlencoded"):
		form, err := url.ParseQuery(string(r.Body))
		if err != nil {
			return nil, nil, err
		}
		for k, vs := range form {
			body[k] = len(vs)
		}
	case strings.HasPrefix(r.ContentType, "application/json"):
		var obj map[string]any
		if err := json.Unmarshal(r.Body, &obj); err != nil {
			return nil, nil, err
		}
		for k := range obj {
			body[k] = 1
		}
	default:
		return nil, nil, fmt.Errorf("unexpected content type %q", r.ContentType)
	}
	return query, body, nil
}

func TestSpecContractRequests(t *testing.T) {
	spec := loadSpec(t)
	calls := runContractWalk(t)
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(contractAccountSid+":"+contractAPIKey))

	covered := map[string]bool{}
	hostsSeen := map[string]map[string]bool{}
	sentKeys := map[string]map[string]bool{}
	for _, call := range calls {
		label := call.Service + "." + call.Method
		for _, r := range call.Reqs {
			pathOnly := strings.TrimPrefix(r.Path, "")
			op := spec.match(r.Method, pathOnly)
			if op == nil {
				t.Errorf("%s: %s %s matches no spec operation", label, r.Method, r.Path)
				continue
			}
			covered[op.ID] = true
			if hostsSeen[op.ID] == nil {
				hostsSeen[op.ID] = map[string]bool{}
			}
			hostsSeen[op.ID][r.Host] = true
			if !op.Hosts[r.Host] {
				t.Errorf("%s: %s routed to %s host; spec tags %v", label, op.ID, r.Host, op.Tags)
			}
			if want := expectedHost(call.Service); r.Host != want {
				t.Errorf("%s: %s sent to %s host, want %s", label, op.ID, r.Host, want)
			}
			if op.NoAuth && r.Auth != "" {
				t.Errorf("%s: %s is security: [] but sent Authorization", label, op.ID)
			}
			if !op.NoAuth && r.Auth != wantAuth {
				t.Errorf("%s: %s Authorization = %q, want Basic credentials", label, op.ID, r.Auth)
			}
			query, body, err := requestKeys(r)
			if err != nil {
				t.Errorf("%s: %s: %v", label, op.ID, err)
				continue
			}
			if sentKeys[op.ID] == nil {
				sentKeys[op.ID] = map[string]bool{}
			}
			for k, vs := range query {
				sentKeys[op.ID]["query:"+k] = true
				p, ok := op.Query[k]
				if !ok {
					t.Errorf("%s: %s sends undeclared query key %q", label, op.ID, k)
					continue
				}
				if len(vs) > 1 && !p.Array {
					t.Errorf("%s: %s repeats scalar query key %q", label, op.ID, k)
				}
			}
			for k, p := range op.Query {
				if p.Required && query.Get(k) == "" {
					t.Errorf("%s: %s omits required query key %q", label, op.ID, k)
				}
			}
			if len(r.Body) == 0 {
				if len(op.Bodies) > 0 && hasRequired(op) {
					t.Errorf("%s: %s sends no body but the spec requires fields", label, op.ID)
				}
				continue
			}
			ct, _, _ := strings.Cut(r.ContentType, ";")
			decl, ok := op.Bodies[ct]
			if !ok {
				t.Errorf("%s: %s sends %s; spec declares %v", label, op.ID, ct, keysOf(op.Bodies))
				continue
			}
			for k, n := range body {
				sentKeys[op.ID]["body:"+k] = true
				p, ok := decl.Props[k]
				if !ok {
					t.Errorf("%s: %s sends undeclared body key %q", label, op.ID, k)
					continue
				}
				if n > 1 && !p.Array {
					t.Errorf("%s: %s repeats scalar body key %q", label, op.ID, k)
				}
			}
			for k := range decl.Required {
				if _, ok := body[k]; !ok {
					t.Errorf("%s: %s omits required body key %q", label, op.ID, k)
				}
			}
		}
	}

	var missing []string
	for _, op := range spec.Ops {
		if covered[op.ID] {
			continue
		}
		if _, ok := uncoverableOps[op.ID]; ok {
			continue
		}
		missing = append(missing, fmt.Sprintf("%s %s (%s)", op.Method, op.Path, op.ID))
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("%d of %d spec operations are not reachable from the SDK:\n  %s", len(missing), len(spec.Ops), strings.Join(missing, "\n  "))
	}
	t.Logf("coverage: %d/%d operations", len(covered), len(spec.Ops))

	// Ops tagged for both products must be exposed on both hosts.
	for _, op := range spec.Ops {
		if len(op.Hosts) < 2 {
			continue
		}
		for h := range op.Hosts {
			if !hostsSeen[op.ID][h] {
				t.Errorf("%s is tagged for the %s host but no SDK method sends it there", op.ID, h)
			}
		}
	}

	// Every declared parameter must be sendable through some SDK method.
	for _, op := range spec.Ops {
		if !covered[op.ID] {
			continue
		}
		var unsendable []string
		for k := range op.Query {
			if !sentKeys[op.ID]["query:"+k] {
				unsendable = append(unsendable, "query "+k)
			}
		}
		for ct, b := range op.Bodies {
			if !sentContentType(calls, spec, op, ct) {
				continue
			}
			for k := range b.Props {
				if !sentKeys[op.ID]["body:"+k] {
					unsendable = append(unsendable, "body "+k)
				}
			}
		}
		sort.Strings(unsendable)
		if len(unsendable) > 0 {
			t.Errorf("%s: SDK cannot send declared params: %s", op.ID, strings.Join(unsendable, ", "))
		}
	}
}

func hasRequired(op *specOp) bool {
	for _, b := range op.Bodies {
		if len(b.Required) > 0 {
			return true
		}
	}
	return false
}

// sentContentType reports whether the SDK uses content type ct for op. Ops
// that accept both form and JSON are only expected to be complete in the
// encoding the SDK sends.
func sentContentType(calls []contractCall, spec *specDoc, op *specOp, ct string) bool {
	for _, c := range calls {
		for _, r := range c.Reqs {
			if spec.match(r.Method, r.Path) != op {
				continue
			}
			got, _, _ := strings.Cut(r.ContentType, ";")
			if got == ct {
				return true
			}
		}
	}
	return false
}

func keysOf[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
