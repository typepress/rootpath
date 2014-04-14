package rootpath

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/go-martini/martini"
	"github.com/typepress/types"
)

const host = "aa.bb"

func TestRootPath(t *testing.T) {

	m := martini.Classic()
	m.Map(http.Dir(""))
	m.Map(types.ContentDir(""))
	m.Map(types.TemplateDir(""))
	rps := make([]RootPath, len(testRoots))
	rss := make([][]string, len(testRoots))
	for i, tr := range testRoots {
		m.Handlers(Handler(0, tr.rp), func(dir http.Dir, cdir types.ContentDir, tdir types.TemplateDir, c martini.Context) {
			if string(dir) != tr.ss[1] || string(cdir) != tr.ss[2] || string(tdir) != tr.ss[3] {
				t.Fatal(tr.rp.Flag, tr.ss, dir, cdir, tdir)
			}
		})

		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "http://"+tr.ss[0], nil)

		if err != nil {
			t.Fatal(err)
		}

		m.ServeHTTP(w, req)

		rps[i] = tr.rp
		rss[i] = tr.ss
		domain := rps[i].Domain
		rps[i].Domain = strconv.Itoa(i) + "." + domain

		for k, s := range rss[i] {
			rss[i][k] = strings.Replace(s, domain, rps[i].Domain, -1)
		}
	}
	var (
		i  int
		ss []string
	)

	m.Handlers(Handler(403, rps...), func(dir http.Dir, cdir types.ContentDir, tdir types.TemplateDir) {

		if string(dir) != rss[i][1] || string(cdir) != rss[i][2] || string(tdir) != rss[i][3] {
			t.Fatal(rps[i].Flag, rps[i].Domain, rps[i].Root, rss[i], dir, cdir, tdir)
		}

	})

	for i, ss = range rss {
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "http://"+ss[0], nil)

		if err != nil {
			t.Fatal(err)
		}

		m.ServeHTTP(w, req)
		if w.Code != 404 && w.Code != 403 || w.Code == 403 && ss[1]+ss[2]+ss[3] != "" {
			t.Fatal(w.Code, rps[i].Flag, ss)
		}
	}
}
func fixPath(ss ...string) []string {
	for i := 1; i < len(ss); i++ {
		if ss[i] != "" {
			ss[i] = filepath.Clean(ss[i])
		}
	}
	return ss
}

type testStruct struct {
	rp RootPath
	ss []string // host,Dir,ContentDir,TemplateDir[,...]
}

// Root/[foo.][example.com]/[category]
var testRoots = []testStruct{
	/*-----------------------equal----------------------------*/
	testStruct{
		RootPath{
			Pattern: "",
			Domain:  host,
			Root:    "equal",
			Flag:    FAll,
		},
		fixPath(
			host,
			"equal/aa.bb/_static", "equal/aa.bb/_content", "equal/aa.bb/_template",
		),
	},
	testStruct{
		RootPath{
			Pattern: "",
			Domain:  host,
			Root:    "equal",
			Flag:    FStatic,
		},
		fixPath(
			host,
			"equal/aa.bb/_static", "", "",
		),
	},
	testStruct{
		RootPath{
			Pattern: "",
			Domain:  host,
			Root:    "equal",
			Flag:    FStatic | FDontJoinCategoryName,
		},
		fixPath(
			host,
			"equal/aa.bb", "", "",
		),
	},
	testStruct{
		RootPath{
			Pattern: "",
			Domain:  host,
			Root:    "equal",
			Flag:    FStatic | FDontJoinDomain,
		},
		fixPath(
			host,
			"equal/_static", "", "",
		),
	},
	testStruct{
		RootPath{
			Pattern: "",
			Domain:  host,
			Root:    "equal",
			Flag:    FStatic | FDontJoinCategoryName | FDontJoinDomain,
		},
		fixPath(
			host,
			"equal", "", "",
		),
	},
	testStruct{
		RootPath{
			Pattern: "",
			Domain:  host,
			Root:    "equal",
			Flag:    FAll,
		},
		fixPath(
			"cc."+host,
			"", "", "",
		),
	},

	/*---------------------------base---------------------------------*/
	testStruct{
		RootPath{
			Pattern: ".",
			Domain:  host,
			Root:    "base",
			Flag:    FStatic | FContent,
		},
		fixPath(
			host,
			"base/aa.bb/_static", "base/aa.bb/_content", "",
		),
	},
	testStruct{
		RootPath{
			Pattern: ".",
			Domain:  host,
			Root:    "base",
			Flag:    FStatic | FContent,
		},
		fixPath(
			"cc."+host,
			"base/aa.bb/_static", "base/aa.bb/_content", "",
		),
	},
	testStruct{
		RootPath{
			Pattern: ".",
			Domain:  host,
			Root:    "base",
			Flag:    FStatic | FDontJoinCategoryName,
		},
		fixPath(
			"cc."+host,
			"base/aa.bb", "", "",
		),
	},
	testStruct{
		RootPath{
			Pattern: ".",
			Domain:  host,
			Root:    "base",
			Flag:    FStatic | FDontJoinDomain,
		},
		fixPath(
			"cc."+host,
			"base/_static", "", "",
		),
	},
	testStruct{
		RootPath{
			Pattern: ".",
			Domain:  host,
			Root:    "base",
			Flag:    FStatic | FDontJoinCategoryName | FDontJoinDomain,
		},
		fixPath(
			"cc."+host,
			"base", "", "",
		),
	},

	/*---------------------------any---------------------------------*/
	testStruct{
		RootPath{
			Pattern: "*",
			Domain:  host,
			Root:    "any",
			Flag:    FStatic | FContent,
		},
		fixPath(
			host,
			"any/aa.bb/_static", "any/aa.bb/_content", "",
		),
	},
	testStruct{
		RootPath{
			Pattern: "*",
			Domain:  host,
			Root:    "any",
			Flag:    FStatic | FContent,
		},
		fixPath(
			"cc."+host,
			"any/cc.aa.bb/_static", "any/cc.aa.bb/_content", "",
		),
	},
	testStruct{
		RootPath{
			Pattern: "*",
			Domain:  host,
			Root:    "any",
			Flag:    FStatic | FDontJoinCategoryName,
		},
		fixPath(
			"cc."+host,
			"any/cc.aa.bb", "", "",
		),
	},
	testStruct{
		RootPath{
			Pattern: "*",
			Domain:  host,
			Root:    "any",
			Flag:    FStatic | FDontJoinDomain,
		},
		fixPath(
			"cc."+host,
			"any/_static", "", "",
		),
	},
	testStruct{
		RootPath{
			Pattern: "*",
			Domain:  host,
			Root:    "any",
			Flag:    FStatic | FDontJoinCategoryName | FDontJoinDomain,
		},
		fixPath(
			"cc."+host,
			"any", "", "",
		),
	},
	/*---------------------------any---------------------------------*/
	testStruct{
		RootPath{
			Pattern:      "*",
			Domain:       host,
			Root:         "any",
			Flag:         FStatic | FContent | FTemplate,
			CategoryName: []string{"", "_posts", "_layouts"},
		},
		fixPath(
			host,
			"any/aa.bb/", "any/aa.bb/_posts", "any/aa.bb/_layouts",
		),
	},
}
