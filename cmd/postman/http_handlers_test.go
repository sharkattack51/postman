package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

//
// Publish
//

func HttpPublishTester(t *testing.T, o Options, r *http.Request, preFn func(w *httptest.ResponseRecorder, r *http.Request), postFn func(w *httptest.ResponseRecorder)) {
	opts = o
	Prepare()

	w := httptest.NewRecorder()
	if preFn != nil {
		preFn(w, r)
	}

	time.Sleep(100 * time.Millisecond) // wait

	// request
	PublishHandler(w, r)

	// response
	require.Equal(t, w.Code, http.StatusOK)

	time.Sleep(100 * time.Millisecond) // wait

	if postFn != nil {
		postFn(w)
	}
}

func HttpPublishPostTester(t *testing.T, o Options, tgt string, j string, preFn func(w *httptest.ResponseRecorder, r *http.Request), postFn func(w *httptest.ResponseRecorder)) {
	form := url.Values{}
	form.Add("json", j)

	r := httptest.NewRequest(http.MethodPost, tgt, strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	HttpPublishTester(t, o, r, preFn, postFn)
}

func TestHttpPublishApi(t *testing.T) {
	// [GET] standard publish as query
	HttpPublishTester(t,
		Options{LogDir: "./log"},
		httptest.NewRequest(http.MethodGet, "/postman/publish", nil),
		func(w *httptest.ResponseRecorder, r *http.Request) {
			// set query
			q := r.URL.Query()
			q.Add("ch", "TEST_CH")
			q.Add("msg", "TEST@MESSAGE")
			q.Add("tag", "TEST@TAG")
			q.Add("ext", "TEST@EXTENTION")
			q.Add("ci", "TEST@INFO")
			r.URL.RawQuery = q.Encode()
		},
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsSuccess(t, w.Body.Bytes())
		})

	// [POST] standard publish as post json
	HttpPublishPostTester(t,
		Options{},
		"/postman/publish",
		`{"channel":"TEST_CH","message":"TEST@MESSAGE","extention":"TEST@EXTENTION"}`,
		nil,
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsSuccess(t, w.Body.Bytes())
		})

	// [GET] publish empty channnel
	HttpPublishTester(t,
		Options{},
		httptest.NewRequest(http.MethodGet, "/postman/publish", nil),
		func(w *httptest.ResponseRecorder, r *http.Request) {
			// set query
			q := r.URL.Query()
			q.Add("channel", "")
			q.Add("message", "TEST@MESSAGE")
			r.URL.RawQuery = q.Encode()
		},
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "publish channel is empty")
		})

	// [GET] publish group channnel
	HttpPublishTester(t,
		Options{},
		httptest.NewRequest(http.MethodGet, "/postman/publish", nil),
		func(w *httptest.ResponseRecorder, r *http.Request) {
			// start server
			s := StartMockServer(t)

			// client connect and subscribe
			RequireConnectAndSubscribe(t, s.URL, "TEST_CH/1", "TEST_CLI")

			// set query
			q := r.URL.Query()
			q.Add("channel", "TEST_CH/*")
			q.Add("message", "TEST@MESSAGE")
			q.Add("client_info", "TEST_CLI")
			r.URL.RawQuery = q.Encode()
		},
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsSuccess(t, w.Body.Bytes())
		})

	// [GET] ip address validation fail
	HttpPublishTester(t,
		Options{IpAddresses: "192.168.0.1"},
		httptest.NewRequest(http.MethodGet, "/postman/publish", nil),
		nil,
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "remote ip blocked")
		})

	// [GET] secure mode fail fail
	HttpPublishTester(t,
		Options{SecureMode: true},
		httptest.NewRequest(http.MethodGet, "/postman/publish", nil),
		nil,
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "security error")
		})
}

//
// Status
//

func HttpStatusTester(t *testing.T, o Options, r *http.Request, preFn func(w *httptest.ResponseRecorder, r *http.Request), postFn func(*httptest.ResponseRecorder)) {
	opts = o
	Prepare()

	w := httptest.NewRecorder()
	if preFn != nil {
		preFn(w, r)
	}

	// request
	if strings.Contains(r.URL.String(), "_pp") {
		StatusPpHandler(w, r)
	} else {
		StatusHandler(w, r)
	}

	// response
	require.Equal(t, w.Code, http.StatusOK)

	if postFn != nil {
		postFn(w)
	}
}

func TestHttpStatusApi(t *testing.T) {
	// [GET] status
	HttpStatusTester(t,
		Options{},
		httptest.NewRequest(http.MethodGet, "/postman/status", nil),
		nil,
		func(w *httptest.ResponseRecorder) {
			var msg StatusMessage
			err := json.Unmarshal(w.Body.Bytes(), &msg)

			require.NoError(t, err)
			require.Equal(t, msg.Version, VERSION)
		})

	// [GET] ip address validation fail
	HttpStatusTester(t,
		Options{IpAddresses: "192.168.0.1"},
		httptest.NewRequest(http.MethodGet, "/postman/status", nil),
		nil,
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "remote ip blocked")
		})

	// [GET] secure mode fail
	HttpStatusTester(t,
		Options{SecureMode: true},
		httptest.NewRequest(http.MethodGet, "/postman/status", nil),
		nil,
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "security error")
		})

	// [GET] status_pp
	HttpStatusTester(t,
		Options{},
		httptest.NewRequest(http.MethodGet, "/postman/status_pp", nil),
		nil,
		func(w *httptest.ResponseRecorder) {
			var msg StatusMessage
			err := json.Unmarshal(w.Body.Bytes(), &msg)

			require.NoError(t, err)
			require.Equal(t, msg.Version, VERSION)
		})

	// [GET] ip address validation fail as pp
	HttpStatusTester(t,
		Options{IpAddresses: "192.168.0.1"},
		httptest.NewRequest(http.MethodGet, "/postman/status_pp", nil),
		nil,
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "remote ip blocked")
		})

	// [GET] secure mode fail as pp
	HttpStatusTester(t,
		Options{SecureMode: true},
		httptest.NewRequest(http.MethodGet, "/postman/status_pp", nil),
		nil,
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "security error")
		})
}

//
// Store
//

func HttpStoreTester(t *testing.T, o Options, r *http.Request, preFn func(w *httptest.ResponseRecorder, r *http.Request), postFn func(*httptest.ResponseRecorder)) {
	opts = o
	Prepare()
	defer kvsDB.Close()

	w := httptest.NewRecorder()
	if preFn != nil {
		preFn(w, r)
	}

	// request
	StoreHandler(w, r)

	// response
	require.Equal(t, w.Code, http.StatusOK)

	if postFn != nil {
		postFn(w)
	}
}

func HttpStorePostTester(t *testing.T, o Options, tgt string, j string, preFn func(w *httptest.ResponseRecorder, r *http.Request), postFn func(w *httptest.ResponseRecorder)) {
	form := url.Values{}
	form.Add("json", j)

	r := httptest.NewRequest(http.MethodPost, tgt, strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	HttpStoreTester(t, o, r, preFn, postFn)
}

func TestHttpStoreApi(t *testing.T) {
	// [GET] store set as query
	HttpStoreTester(t,
		Options{UseStoreApi: true},
		httptest.NewRequest(http.MethodGet, "/postman/store", nil),
		func(w *httptest.ResponseRecorder, r *http.Request) {
			// set query
			q := r.URL.Query()
			q.Add("command", "SET")
			q.Add("key", "TEST#KEY")
			q.Add("value", "1234")
			r.URL.RawQuery = q.Encode()
		},
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsSuccess(t, w.Body.Bytes())
		})

	// [GET] store get as query
	HttpStoreTester(t,
		Options{UseStoreApi: true},
		httptest.NewRequest(http.MethodGet, "/postman/store", nil),
		func(w *httptest.ResponseRecorder, r *http.Request) {
			// set query
			q := r.URL.Query()
			q.Add("cmd", "get")
			q.Add("key", "TEST#KEY")
			r.URL.RawQuery = q.Encode()
		},
		func(w *httptest.ResponseRecorder) {
			var msg ResultMessage
			err := json.Unmarshal(w.Body.Bytes(), &msg)

			require.NoError(t, err)
			require.Equal(t, msg.Result, "1234")
		})

	// [POST] store has as post json
	HttpStorePostTester(t,
		Options{UseStoreApi: true},
		"/postman/store",
		`{"cmd":"HAS","key":"TEST#KEY"}`,
		nil,
		func(w *httptest.ResponseRecorder) {
			var msg ResultMessage
			err := json.Unmarshal(w.Body.Bytes(), &msg)

			require.NoError(t, err)
			require.Equal(t, msg.Result, "true")
		})

	// [POST] store del as post json
	HttpStorePostTester(t,
		Options{UseStoreApi: true},
		"/postman/store",
		`{"cmd":"del","key":"TEST#KEY"}`,
		nil,
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsSuccess(t, w.Body.Bytes())
		})

	// [GET] store command not found
	HttpStoreTester(t,
		Options{UseStoreApi: true},
		httptest.NewRequest(http.MethodGet, "/postman/store", nil),
		func(w *httptest.ResponseRecorder, r *http.Request) {
			// set query
			q := r.URL.Query()
			q.Add("cmd", "gettt")
			q.Add("key", "TEST#KEY")
			r.URL.RawQuery = q.Encode()
		},
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "store command not found")
		})

	// [GET] store key is empty
	HttpStoreTester(t,
		Options{UseStoreApi: true},
		httptest.NewRequest(http.MethodGet, "/postman/store", nil),
		func(w *httptest.ResponseRecorder, r *http.Request) {
			// set query
			q := r.URL.Query()
			q.Add("cmd", "get")
			r.URL.RawQuery = q.Encode()
		},
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "store key is empty")
		})

	// [GET] ip address validation fail
	HttpStoreTester(t,
		Options{UseStoreApi: true, IpAddresses: "192.168.0.1"},
		httptest.NewRequest(http.MethodGet, "/postman/store", nil),
		nil,
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "remote ip blocked")
		})

	// [POST] secure mode fail as post json
	HttpStorePostTester(t,
		Options{UseStoreApi: true, SecureMode: true},
		"/postman/store",
		`{"cmd":"set","key":"@@@", "tkn":"absde"}`,
		nil,
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "security error")
		})

	// [GET] disable store api
	HttpStoreTester(t,
		Options{},
		httptest.NewRequest(http.MethodGet, "/postman/store", nil),
		nil,
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "key-value store api is disable")
		})

	// [POST] store get no key
	HttpStorePostTester(t,
		Options{UseStoreApi: true},
		"/postman/store",
		`{"cmd":"get","key":"@@@"}`,
		nil,
		func(w *httptest.ResponseRecorder) {
			var msg ResultMessage
			err := json.Unmarshal(w.Body.Bytes(), &msg)

			require.NoError(t, err)
			require.Equal(t, msg.Result, "")
		})

	// [POST] store set error
	HttpStorePostTester(t,
		Options{UseStoreApi: true},
		"/postman/store",
		`{"cmd":"set","key":"@@@"}`,
		func(w *httptest.ResponseRecorder, r *http.Request) {
			kvsDB.Close()
		},
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "leveldb: closed")
		})

	// [GET] store has error
	HttpStoreTester(t,
		Options{UseStoreApi: true},
		httptest.NewRequest(http.MethodGet, "/postman/store", nil),
		func(w *httptest.ResponseRecorder, r *http.Request) {
			kvsDB.Close()

			// set query
			q := r.URL.Query()
			q.Add("cmd", "HAS")
			q.Add("key", "@@@")
			r.URL.RawQuery = q.Encode()
		},
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "leveldb: closed")
		})

	// [GET] store del error
	HttpStoreTester(t,
		Options{UseStoreApi: true},
		httptest.NewRequest(http.MethodGet, "/postman/store", nil),
		func(w *httptest.ResponseRecorder, r *http.Request) {
			kvsDB.Close()

			// set query
			q := r.URL.Query()
			q.Add("cmd", "DEL")
			q.Add("key", "@@@")
			r.URL.RawQuery = q.Encode()
		},
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "leveldb: closed")
		})

	// [GET] store command is empty
	HttpStoreTester(t,
		Options{UseStoreApi: true},
		httptest.NewRequest(http.MethodGet, "/postman/store", nil),
		nil,
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "store command is empty")
		})
}

//
// File
//

func HttpFileTester(t *testing.T, o Options, r *http.Request, preFn func(w *httptest.ResponseRecorder, r *http.Request), postFn func(*httptest.ResponseRecorder)) {
	opts = o
	Prepare()

	w := httptest.NewRecorder()
	if preFn != nil {
		preFn(w, r)
	}

	// request
	FileHandler(w, r)

	// response
	require.Equal(t, w.Code, http.StatusOK)

	if postFn != nil {
		postFn(w)
	}
}

func HttpFilePostTester(t *testing.T, o Options, tgt string, src string, dst string, preFn func(w *httptest.ResponseRecorder, r *http.Request), postFn func(*httptest.ResponseRecorder)) {
	var b bytes.Buffer
	ctype := ""

	if src != "" && dst != "" {
		if IsExist(filepath.Join(SERVE_FILES_DIR, dst)) {
			os.Remove(filepath.Join(SERVE_FILES_DIR, dst))
		}

		mpw := multipart.NewWriter(&b)
		fw, _ := mpw.CreateFormFile("file", dst)
		f, _ := os.Open(filepath.Join(SERVE_FILES_DIR, src))
		io.Copy(fw, f)
		f.Close()
		mpw.Close()

		ctype = mpw.FormDataContentType()
	}

	r := httptest.NewRequest(http.MethodPost, tgt, &b)
	r.Header.Set("Content-Type", ctype)

	HttpFileTester(t, o, r, preFn, postFn)
}

func TestHttpFileApi(t *testing.T) {
	// [GET] standard get file
	HttpFileTester(t,
		Options{UseFileApi: true},
		httptest.NewRequest(http.MethodGet, "/postman/file/test.txt", nil),
		nil,
		func(w *httptest.ResponseRecorder) {
			require.Equal(t, w.Body.String(), "test")
		})

	// [GET] ip address validation fail
	HttpFileTester(t,
		Options{UseFileApi: true, IpAddresses: "192.168.0.1"},
		httptest.NewRequest(http.MethodGet, "/postman/file/test.txt", nil),
		nil,
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "remote ip blocked")
		})

	// [GET] secure mode fail
	HttpFileTester(t,
		Options{UseFileApi: true, SecureMode: true},
		httptest.NewRequest(http.MethodGet, "/postman/file/test.txt?token=TOKEN", nil),
		nil,
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "security error")
		})

	// [GET] disable file api
	HttpFileTester(t,
		Options{},
		httptest.NewRequest(http.MethodGet, "/postman/file/test.txt", nil),
		nil,
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "file server api is disable")
		})

	// [GET] not found file serve directory
	HttpFileTester(t,
		Options{UseFileApi: true},
		httptest.NewRequest(http.MethodGet, "/postman/file/test.txt", nil),
		func(w *httptest.ResponseRecorder, r *http.Request) {
			os.Rename(SERVE_FILES_DIR, "_"+SERVE_FILES_DIR)
		},
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), fmt.Sprintf("directory not found \"%s\"", SERVE_FILES_DIR)) // fail
			os.Rename("_"+SERVE_FILES_DIR, SERVE_FILES_DIR)
		})

	// [GET] not found file serve file
	HttpFileTester(t,
		Options{UseFileApi: true},
		httptest.NewRequest(http.MethodGet, "/postman/file/notfound.txt", nil),
		nil,
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "file not found \"notfound.txt\"")
		})

	// [GET] index.html access
	HttpFileTester(t,
		Options{UseFileApi: true},
		httptest.NewRequest(http.MethodGet, "/postman/file", nil),
		nil,
		func(w *httptest.ResponseRecorder) {
			require.Contains(t, w.Body.String(), "<!DOCTYPE html>") // index.html
		})

	// [POST] standard post file
	HttpFilePostTester(t,
		Options{UseFileApi: true},
		"/postman/file",
		"test.txt",
		"test2.txt",
		nil,
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsSuccess(t, w.Body.Bytes())

			b1, _ := os.ReadFile(filepath.Join(SERVE_FILES_DIR, "test.txt"))
			b2, _ := os.ReadFile(filepath.Join(SERVE_FILES_DIR, "test2.txt"))
			require.Equal(t, string(b1), string(b2))

			os.Remove(filepath.Join(SERVE_FILES_DIR, "test2.txt"))
		})

	// [POST] no form
	HttpFilePostTester(t,
		Options{UseFileApi: true},
		"/postman/file",
		"",
		"",
		nil,
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "no form file data")
		})

	// [POST] write file error
	HttpFilePostTester(t,
		Options{UseFileApi: true},
		"/postman/file",
		"test.txt",
		"test2.txt",
		func(w *httptest.ResponseRecorder, r *http.Request) {
			os.Chmod(SERVE_FILES_DIR, fs.FileMode(0155))
		},
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "could not write file \"test2.txt\"")
			os.Chmod(SERVE_FILES_DIR, fs.FileMode(0755))
		})
}

//
// Plugin
//

func HttpPluginTester(t *testing.T, o Options, r *http.Request, preFn func(w *httptest.ResponseRecorder, r *http.Request), postFn func(*httptest.ResponseRecorder)) {
	opts = o
	Prepare()

	w := httptest.NewRecorder()
	if preFn != nil {
		preFn(w, r)
	}

	// request
	PluginHandler(w, r)

	// response
	require.Equal(t, w.Code, http.StatusOK)

	if postFn != nil {
		postFn(w)
	}
}

func HttpPluginPostTester(t *testing.T, o Options, tgt string, j string, preFn func(w *httptest.ResponseRecorder, r *http.Request), postFn func(*httptest.ResponseRecorder)) {
	form := url.Values{}
	form.Add("json", j)

	r := httptest.NewRequest(http.MethodPost, tgt, strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	HttpPluginTester(t, o, r, preFn, postFn)
}

func TestHttpPluginApi(t *testing.T) {
	// [GET] call plgin as query
	HttpPluginTester(t,
		Options{UsePluginApi: true},
		httptest.NewRequest(http.MethodGet, "/postman/plugin", nil),
		func(w *httptest.ResponseRecorder, r *http.Request) {
			// set query
			q := r.URL.Query()
			q.Add("cmd", "example_cli") // ls -a
			r.URL.RawQuery = q.Encode()
		},
		func(w *httptest.ResponseRecorder) {
			require.Equal(t, string(w.Body.String()[0]), ".")
		})

	// [POST] call plgin as post json
	HttpPluginPostTester(t,
		Options{UsePluginApi: true},
		"/postman/plugin",
		`{"cmd":"example_cli"}`,
		nil,
		func(w *httptest.ResponseRecorder) {
			require.Equal(t, string(w.Body.String()[0]), ".")
		})

	// [GET] call plgin command not found
	HttpPluginTester(t,
		Options{UsePluginApi: true},
		httptest.NewRequest(http.MethodGet, "/postman/plugin", nil),
		func(w *httptest.ResponseRecorder, r *http.Request) {
			// set query
			q := r.URL.Query()
			q.Add("cmd", "notfound")
			r.URL.RawQuery = q.Encode()
		},
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "plugin command not found")
		})

	// [GET] ip address validation fail
	HttpPluginTester(t,
		Options{UsePluginApi: true, IpAddresses: "192.168.0.1"},
		httptest.NewRequest(http.MethodGet, "/postman/plugin", nil),
		nil,
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "remote ip blocked")
		})

	// [POST] secure mode fail as post json
	HttpPluginPostTester(t,
		Options{UsePluginApi: true, SecureMode: true},
		"/postman/plugin",
		`{"cmd":"example_cli","tkn":"absde"}`,
		nil,
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "security error")
		})

	// [GET] disable plugin api
	HttpPluginTester(t,
		Options{},
		httptest.NewRequest(http.MethodGet, "/postman/plugin", nil),
		nil,
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "plugin api is disable")
		})

	// [GET] plugin command is empty
	HttpPluginTester(t,
		Options{UsePluginApi: true},
		httptest.NewRequest(http.MethodGet, "/postman/plugin", nil),
		nil,
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "plugin command is empty")
		})

	// [POST] plugin load error
	b := make([]byte, 1024)
	HttpPluginPostTester(t,
		Options{UsePluginApi: true},
		"/postman/plugin",
		`{"command":"example_cli"}`,
		func(w *httptest.ResponseRecorder, r *http.Request) {
			b, _ = os.ReadFile(filepath.Join(PLUGIN_DIR, PLUGIN_JSON))
			os.WriteFile(filepath.Join(PLUGIN_DIR, PLUGIN_JSON), []byte("{"), 0666) // break json
		},
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "plugin can't loaded")
			os.WriteFile(filepath.Join(PLUGIN_DIR, PLUGIN_JSON), b, 0666) // repair json
		})

	// [GET] plugin dir not found
	HttpPluginTester(t,
		Options{UsePluginApi: true},
		httptest.NewRequest(http.MethodGet, "/postman/plugin", nil),
		func(w *httptest.ResponseRecorder, r *http.Request) {
			os.Rename(PLUGIN_DIR, PLUGIN_DIR+"_")

			// set query
			q := r.URL.Query()
			q.Add("cmd", "test")
			r.URL.RawQuery = q.Encode()
		},
		func(w *httptest.ResponseRecorder) {
			RequireResponseIsFail(t, w.Body.Bytes(), "plugin command not found")

			os.RemoveAll(PLUGIN_DIR)
			os.Rename(PLUGIN_DIR+"_", PLUGIN_DIR)
		})
}
