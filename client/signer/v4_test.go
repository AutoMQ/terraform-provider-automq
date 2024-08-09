package signer

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func epochTime() time.Time { return time.Unix(0, 0) }

func TestStripExcessHeaders(t *testing.T) {
	vals := []string{
		"",
		"123",
		"1 2 3",
		"1 2 3 ",
		"  1 2 3",
		"1  2 3",
		"1  23",
		"1  2  3",
		"1  2  ",
		" 1  2  ",
		"12   3",
		"12   3   1",
		"12           3     1",
		"12     3       1abc123",
	}

	expected := []string{
		"",
		"123",
		"1 2 3",
		"1 2 3",
		"1 2 3",
		"1 2 3",
		"1 23",
		"1 2 3",
		"1 2",
		"1 2",
		"12 3",
		"12 3 1",
		"12 3 1",
		"12 3 1abc123",
	}

	stripExcessSpaces(vals)
	for i := 0; i < len(vals); i++ {
		if e, a := expected[i], vals[i]; e != a {
			t.Errorf("%d, expect %v, got %v", i, e, a)
		}
	}
}

func buildRequest(serviceName, region, body string) (*http.Request, io.ReadSeeker) {
	reader := strings.NewReader(body)
	return buildRequestWithBodyReader(serviceName, region, reader)
}

func buildRequestReaderSeeker(serviceName, region, body string) (*http.Request, io.ReadSeeker) {
	reader := &readerSeekerWrapper{strings.NewReader(body)}
	return buildRequestWithBodyReader(serviceName, region, reader)
}

func buildRequestWithBodyReader(serviceName, region string, body io.Reader) (*http.Request, io.ReadSeeker) {
	var bodyLen int

	type lenner interface {
		Len() int
	}
	if lr, ok := body.(lenner); ok {
		bodyLen = lr.Len()
	}

	endpoint := "https://" + serviceName + "." + region + ".amazonaws.com"
	req, _ := http.NewRequest("POST", endpoint, body)
	req.URL.Opaque = "//example.org/bucket/key-._~,!@#$%^&*()"
	req.Header.Set("X-Automq-Target", "prefix.Operation")
	req.Header.Set("Content-Type", "application/x-automq-json-1.0")

	if bodyLen > 0 {
		req.Header.Set("Content-Length", strconv.Itoa(bodyLen))
	}

	req.Header.Set("X-Automq-Meta-Other-Header", "some-value=!@#$%^&* (+)")
	req.Header.Add("X-Automq-Meta-Other-Header_With_Underscore", "some-value=!@#$%^&* (+)")
	req.Header.Add("X-Automq-Meta-Other-Header_With_Underscore", "some-value=!@#$%^&* (+)")

	var seeker io.ReadSeeker
	if sr, ok := body.(io.ReadSeeker); ok {
		seeker = sr
	}

	return req, seeker
}

func buildSigner() Signer {
	return Signer{
		Credential: Credentials{AccessKeyID: "AKID", SecretAccessKey: "SECRET"},
	}
}

func TestPresignRequest(t *testing.T) {
	req, body := buildRequest("dynamodb", "private", "{}")

	signer := buildSigner()
	_, err := signer.Presign(req, body, "dynamodb", "private", 300*time.Second, epochTime())
	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}

	expectedDate := "19700101T000000Z"
	expectedHeaders := "content-length;content-type;host;x-automq-meta-other-header;x-automq-meta-other-header_with_underscore"
	expectedSig := "6760ea122b83b0589a2d3e9fc8b9e56786dc69397cac7fbd6a8c4d6930040745"
	expectedCred := "AKID/19700101/private/dynamodb/cmp_request"
	expectedTarget := "prefix.Operation"

	q := req.URL.Query()
	if e, a := expectedSig, q.Get("X-Automq-Signature"); e != a {
		t.Errorf("expect %v, got %v", e, a)
	}
	if e, a := expectedCred, q.Get("X-Automq-Credential"); e != a {
		t.Errorf("expect %v, got %v", e, a)
	}
	if e, a := expectedHeaders, q.Get("X-Automq-SignedHeaders"); e != a {
		t.Errorf("expect %v, got %v", e, a)
	}
	if e, a := expectedDate, q.Get("X-Automq-Date"); e != a {
		t.Errorf("expect %v, got %v", e, a)
	}
	if a := q.Get("X-Automq-Meta-Other-Header"); len(a) != 0 {
		t.Errorf("expect %v to be empty", a)
	}
	if e, a := expectedTarget, q.Get("X-Automq-Target"); e != a {
		t.Errorf("expect %v, got %v", e, a)
	}
}

func TestPresignBodyWithArrayRequest(t *testing.T) {
	req, body := buildRequest("dynamodb", "private", "{}")
	req.URL.RawQuery = "Foo=z&Foo=o&Foo=m&Foo=a"

	signer := buildSigner()
	_, err := signer.Presign(req, body, "dynamodb", "private", 300*time.Second, epochTime())
	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}

	expectedDate := "19700101T000000Z"
	expectedHeaders := "content-length;content-type;host;x-automq-meta-other-header;x-automq-meta-other-header_with_underscore"
	expectedSig := "0835dd81aaea2bf9d8d2d6b57552beef4afdf3b20d7f46e7ed3ebc946884be94"
	expectedCred := "AKID/19700101/private/dynamodb/cmp_request"
	expectedTarget := "prefix.Operation"

	q := req.URL.Query()
	if e, a := expectedSig, q.Get("X-Automq-Signature"); e != a {
		t.Errorf("expect %v, got %v", e, a)
	}
	if e, a := expectedCred, q.Get("X-Automq-Credential"); e != a {
		t.Errorf("expect %v, got %v", e, a)
	}
	if e, a := expectedHeaders, q.Get("X-Automq-SignedHeaders"); e != a {
		t.Errorf("expect %v, got %v", e, a)
	}
	if e, a := expectedDate, q.Get("X-Automq-Date"); e != a {
		t.Errorf("expect %v, got %v", e, a)
	}
	if a := q.Get("X-Automq-Meta-Other-Header"); len(a) != 0 {
		t.Errorf("expect %v to be empty, was not", a)
	}
	if e, a := expectedTarget, q.Get("X-Automq-Target"); e != a {
		t.Errorf("expect %v, got %v", e, a)
	}
}

func TestSignRequest(t *testing.T) {
	req, body := buildRequest("dynamodb", "private", "{}")
	signer := buildSigner()
	_, err := signer.Sign(req, body, "dynamodb", "private", epochTime())
	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}

	expectedDate := "19700101T000000Z"
	expectedSig := "AUTOMQ-HMAC-SHA256 Credential=AKID/19700101/private/dynamodb/cmp_request, SignedHeaders=content-length;content-type;host;x-automq-date;x-automq-meta-other-header;x-automq-meta-other-header_with_underscore;x-automq-target, Signature=b5ab44d63004fb6968f02bc270d1c6a0b90ff4216b70fec8f0cc30364aef8040"

	q := req.Header
	if e, a := expectedSig, q.Get("Authorization"); e != a {
		t.Errorf("expect\n%v\nactual\n%v\n", e, a)
	}
	if e, a := expectedDate, q.Get("X-Automq-Date"); e != a {
		t.Errorf("expect\n%v\nactual\n%v\n", e, a)
	}
}

func TestPresign_UnsignedPayload(t *testing.T) {
	req, body := buildRequest("service-name", "private", "hello")
	signer := buildSigner()
	signer.UnsignedPayload = true
	_, err := signer.Presign(req, body, "service-name", "private", 5*time.Minute, time.Now())
	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}
	hash := req.Header.Get("X-Automq-Content-Sha256")
	if e, a := "UNSIGNED-PAYLOAD", hash; e != a {
		t.Errorf("expect %v, got %v", e, a)
	}
}

func TestSignUnsignedPayloadUnseekableBody(t *testing.T) {
	req, body := buildRequestWithBodyReader("mock-service", "mock-region", bytes.NewBuffer([]byte("hello")))

	signer := buildSigner()
	signer.UnsignedPayload = true

	_, err := signer.Sign(req, body, "mock-service", "mock-region", time.Now())
	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}

	hash := req.Header.Get("X-Automq-Content-Sha256")
	if e, a := "UNSIGNED-PAYLOAD", hash; e != a {
		t.Errorf("expect %v, got %v", e, a)
	}
}

func TestSignPreComputedHashUnseekableBody(t *testing.T) {
	req, body := buildRequestWithBodyReader("mock-service", "mock-region", bytes.NewBuffer([]byte("hello")))

	signer := buildSigner()

	req.Header.Set("X-Automq-Content-Sha256", "some-content-sha256")
	_, err := signer.Sign(req, body, "mock-service", "mock-region", time.Now())
	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}

	hash := req.Header.Get("X-Automq-Content-Sha256")
	if e, a := "some-content-sha256", hash; e != a {
		t.Errorf("expect %v, got %v", e, a)
	}
}

func TestSignPrecomputedBodyChecksum(t *testing.T) {
	req, body := buildRequest("dynamodb", "private", "hello")
	req.Header.Set("X-Automq-Content-Sha256", "PRECOMPUTED")
	signer := buildSigner()
	_, err := signer.Sign(req, body, "dynamodb", "private", time.Now())
	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}
	hash := req.Header.Get("X-Automq-Content-Sha256")
	if e, a := "PRECOMPUTED", hash; e != a {
		t.Errorf("expect %v, got %v", e, a)
	}
}

func TestSignWithRequestBody(t *testing.T) {
	creds := Credentials{AccessKeyID: "123", SecretAccessKey: "321"}
	signer := NewSigner(creds)

	expectBody := []byte("abc123")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			t.Errorf("expect no error, got %v", err)
		}
		if e, a := expectBody, b; !reflect.DeepEqual(e, a) {
			t.Errorf("expect %v, got %v", e, a)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req, err := http.NewRequest("POST", server.URL, nil)
	if err != nil {
		t.Errorf("expect not no error, got %v", err)
	}

	_, err = signer.Sign(req, bytes.NewReader(expectBody), "service", "region", time.Now())
	if err != nil {
		t.Errorf("expect not no error, got %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("expect not no error, got %v", err)
	}
	if e, a := http.StatusOK, resp.StatusCode; e != a {
		t.Errorf("expect %v, got %v", e, a)
	}
}

func TestSignWithRequestBody_Overwrite(t *testing.T) {
	creds := Credentials{AccessKeyID: "123", SecretAccessKey: "321"}
	signer := NewSigner(creds)

	var expectBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			t.Errorf("expect not no error, got %v", err)
		}
		if e, a := len(expectBody), len(b); e != a {
			t.Errorf("expect %v, got %v", e, a)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL, strings.NewReader("invalid body"))
	if err != nil {
		t.Errorf("expect not no error, got %v", err)
	}

	_, err = signer.Sign(req, nil, "service", "region", time.Now())
	req.ContentLength = 0

	if err != nil {
		t.Errorf("expect not no error, got %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("expect not no error, got %v", err)
	}
	if e, a := http.StatusOK, resp.StatusCode; e != a {
		t.Errorf("expect %v, got %v", e, a)
	}
}

func TestBuildCanonicalRequest(t *testing.T) {
	req, body := buildRequest("dynamodb", "private", "{}")
	req.URL.RawQuery = "Foo=z&Foo=o&Foo=m&Foo=a"
	ctx := &signingCtx{
		ServiceName: "dynamodb",
		Region:      "private",
		Request:     req,
		Body:        body,
		Query:       req.URL.Query(),
		Time:        time.Now(),
		ExpireTime:  5 * time.Second,
	}

	ctx.buildCanonicalString()
	expected := "https://example.org/bucket/key-._~,!@#$%^&*()?Foo=z&Foo=o&Foo=m&Foo=a"
	if e, a := expected, ctx.Request.URL.String(); e != a {
		t.Errorf("expect %v, got %v", e, a)
	}
}

func TestSignWithBody_ReplaceRequestBody(t *testing.T) {
	creds := Credentials{AccessKeyID: "123", SecretAccessKey: "321"}
	req, seekerBody := buildRequest("dynamodb", "private", "{}")
	req.Body = io.NopCloser(bytes.NewReader([]byte{}))

	s := NewSigner(creds)
	origBody := req.Body

	_, err := s.Sign(req, seekerBody, "dynamodb", "private", time.Now())
	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}

	if req.Body == origBody {
		t.Errorf("expeect request body to not be origBody")
	}

	if req.Body == nil {
		t.Errorf("expect request body to be changed but was nil")
	}
}

func TestSignWithBody_NoReplaceRequestBody(t *testing.T) {
	creds := Credentials{AccessKeyID: "123", SecretAccessKey: "321"}
	req, seekerBody := buildRequest("dynamodb", "private", "{}")
	req.Body = io.NopCloser(bytes.NewReader([]byte{}))

	s := NewSigner(creds, func(signer *Signer) {
		signer.DisableRequestBodyOverwrite = true
	})

	origBody := req.Body

	_, err := s.Sign(req, seekerBody, "dynamodb", "private", time.Now())
	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}

	if req.Body != origBody {
		t.Errorf("expect request body to not be chagned")
	}
}

func TestRequestHost(t *testing.T) {
	req, body := buildRequest("dynamodb", "private", "{}")
	req.URL.RawQuery = "Foo=z&Foo=o&Foo=m&Foo=a"
	req.Host = "myhost"
	ctx := &signingCtx{
		ServiceName: "dynamodb",
		Region:      "private",
		Request:     req,
		Body:        body,
		Query:       req.URL.Query(),
		Time:        time.Now(),
		ExpireTime:  5 * time.Second,
	}

	ctx.buildCanonicalHeaders(ignoredHeaders, ctx.Request.Header)
	if !strings.Contains(ctx.canonicalHeaders, "host:"+req.Host) {
		t.Errorf("canonical host header invalid")
	}
}

func TestSign_buildCanonicalHeaders(t *testing.T) {
	serviceName := "mockAPI"
	region := "mock-region"
	endpoint := "https://" + serviceName + "." + region + ".amazonaws.com"

	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		t.Fatalf("failed to create request, %v", err)
	}

	req.Header.Set("FooInnerSpace", "   inner      space    ")
	req.Header.Set("FooLeadingSpace", "    leading-space")
	req.Header.Add("FooMultipleSpace", "no-space")
	req.Header.Add("FooMultipleSpace", "\ttab-space")
	req.Header.Add("FooMultipleSpace", "trailing-space    ")
	req.Header.Set("FooNoSpace", "no-space")
	req.Header.Set("FooTabSpace", "\ttab-space\t")
	req.Header.Set("FooTrailingSpace", "trailing-space    ")
	req.Header.Set("FooWrappedSpace", "   wrapped-space    ")

	ctx := &signingCtx{
		ServiceName: serviceName,
		Region:      region,
		Request:     req,
		Body:        nil,
		Query:       req.URL.Query(),
		Time:        time.Now(),
		ExpireTime:  5 * time.Second,
	}

	ctx.buildCanonicalHeaders(ignoredHeaders, ctx.Request.Header)

	expectCanonicalHeaders := strings.Join([]string{
		`fooinnerspace:inner space`,
		`fooleadingspace:leading-space`,
		`foomultiplespace:no-space,tab-space,trailing-space`,
		`foonospace:no-space`,
		`footabspace:tab-space`,
		`footrailingspace:trailing-space`,
		`foowrappedspace:wrapped-space`,
		`host:mockAPI.mock-region.amazonaws.com`,
	}, "\n")
	if e, a := expectCanonicalHeaders, ctx.canonicalHeaders; e != a {
		t.Errorf("expect:\n%s\n\nactual:\n%s", e, a)
	}
}

func BenchmarkPresignRequest(b *testing.B) {
	signer := buildSigner()
	req, body := buildRequestReaderSeeker("dynamodb", "private", "{}")
	for i := 0; i < b.N; i++ {
		_, err := signer.Presign(req, body, "dynamodb", "private", 300*time.Second, time.Now())
		if err != nil {
			b.Fatalf("expect no error, got %v", err)
		}
	}
}

func BenchmarkSignRequest(b *testing.B) {
	signer := buildSigner()
	req, body := buildRequestReaderSeeker("dynamodb", "private", "{}")
	for i := 0; i < b.N; i++ {
		_, err := signer.Sign(req, body, "dynamodb", "private", time.Now())
		if err != nil {
			b.Fatalf("expect no error, got %v", err)
		}
	}
}

var stripExcessSpaceCases = []string{
	`AUTOMQ-HMAC-SHA256 Credential=AKIDFAKEIDFAKEID/20160628/us-west-2/s3/cmp_request, SignedHeaders=host;x-automq-date, Signature=1234567890abcdef1234567890abcdef1234567890abcdef`,
	`123   321   123   321`,
	`   123   321   123   321   `,
	`   123    321    123          321   `,
	"123",
	"1 2 3",
	"  1 2 3",
	"1  2 3",
	"1  23",
	"1  2  3",
	"1  2  ",
	" 1  2  ",
	"12   3",
	"12   3   1",
	"12           3     1",
	"12     3       1abc123",
}

func BenchmarkStripExcessSpaces(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// Make sure to start with a copy of the cases
		cases := append([]string{}, stripExcessSpaceCases...)
		stripExcessSpaces(cases)
	}
}

// readerSeekerWrapper mimics the interface provided by request.offsetReader
type readerSeekerWrapper struct {
	r *strings.Reader
}

func (r *readerSeekerWrapper) Read(p []byte) (n int, err error) {
	return r.r.Read(p)
}

func (r *readerSeekerWrapper) Seek(offset int64, whence int) (int64, error) {
	return r.r.Seek(offset, whence)
}

func (r *readerSeekerWrapper) Len() int {
	return r.r.Len()
}
