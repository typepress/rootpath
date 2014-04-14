// domain rootpath Handler support for Martini and TypePress

// 根据域名匹配相关目录
package rootpath

import (
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/go-martini/martini"
	"github.com/typepress/types"
)

const (
	// flags of category for RootPath.Flag
	FStatic        = 1 << iota // category name is _static
	FContent                   // category name is _content
	FTemplate                  // category name is _template
	fCategoryCount = iota
	FAll           = FStatic | FContent | FTemplate

	// flags of target rootpath for RootPath.Flag
	FDontJoinCategoryName = 0x40000000
	FDontJoinDomain       = 0x80000000
)

var (
	pathedType reflect.Type = reflect.TypeOf(rootPathed(true))

	categoryType = [fCategoryCount]func(string) interface{}{
		func(s string) interface{} { return http.Dir(s) },
		func(s string) interface{} { return types.ContentDir(s) },
		func(s string) interface{} { return types.TemplateDir(s) },
	}

	categoryFlag = [fCategoryCount]int{
		FStatic, FContent, FTemplate,
	}

	categoryName = [fCategoryCount]string{
		"_static", "_content", "_template",
	}
)

type rootPathed bool

/**
  返回多域名路径设置 handler. 依据 roots 匹配 Request.Host.
  匹配成功设置相应路径, 无匹配时依据 statusCode 进行操作.
  statusCode:
    - 0 无操作
    - 其他 WriteHeader(statusCode)
*/
func Handler(statusCode int, root ...RootPath) martini.Handler {
	lock := sync.RWMutex{}
	cacheDir := map[string]int{} // save file exist status for "?" pattern

	// clone
	root = append([]RootPath{}, root...)
	for i, _ := range root {
		root[i].CategoryName = append([]string{}, root[i].CategoryName...)
		l := len(root[i].CategoryName)
		if l < fCategoryCount {
			root[i].CategoryName = append(root[i].CategoryName, categoryName[l:]...)
		}
	}

	return func(res http.ResponseWriter, req *http.Request, c martini.Context) {
		v := c.Get(pathedType)
		var pathed bool
		if v.IsValid() && v.Bool() {
			return
		}

		host := strings.SplitN(req.Host, ":", 2)[0]

		for _, rp := range root {
			ok, prefix := rp.Match(host)
			if !ok {
				continue
			}

			for i := 0; i < fCategoryCount; i++ {
				if rp.Flag&categoryFlag[i] == 0 {
					continue
				}

				var domain, category string
				var exist int
				if 0 == FDontJoinDomain&rp.Flag {
					if len(prefix) == 0 {
						domain = rp.Domain
					} else {
						domain = prefix + "." + rp.Domain
					}
				}

				if 0 == FDontJoinCategoryName&rp.Flag {
					category = rp.CategoryName[i]
				}

				dir := filepath.Join(rp.Root, domain, category)

				if rp.Pattern == "?" {

					lock.RLock()
					exist = cacheDir[dir]
					lock.RUnlock()

					if exist == 0 {
						_, err := os.Stat(dir)

						if err == nil {
							exist = 1
						} else {
							exist = -1
						}

						lock.Lock()
						cacheDir[dir] = exist // dir exist
						lock.Unlock()
					}

					if exist == -1 {
						if 0 == FDontJoinDomain&rp.Flag {
							dir = filepath.Join(rp.Root, rp.Domain, category)
						} else {
							dir = filepath.Join(rp.Root, category)
						}
					}

				}
				c.Map(categoryType[i](dir))
				pathed = true
			}
			if pathed {
				c.Map(rootPathed(true))
				return
			}
		}

		if statusCode != 0 {
			res.WriteHeader(statusCode)
			return
		}
	}
}

/**
  RootPath 使用简单的规则来确定域名根目录

  Pattern: 域名匹配规则, "?" 规则多了一次目录检查, 是否 join example.com 部分由 FDontJoinDomain 标记决定

	- ""  完全匹配             example.com , Root/[example.com]/[category]
	- "." 泛域名相同目录 [foo.]example.com , Root/[example.com]/[category]
	- "*" 泛域名独立目录 [foo.]example.com , Root/[foo.][example.com]/[category]
	- "?" 泛域名泛目录   [foo.]example.com , Root/[foo.][example.com]/[category] 或 Root/[example.com]/[category]

  Domain: 域名
  Root:   基本目录
  Flag:   根目录的分类标记, 参见 FStatic 等常量, 0 特指所有类型
  CategoryName: 自定义 category 的名字, 顺序对应 static,content,template. 缺省使用 "_static","_content","_template"
*/
type RootPath struct {
	Pattern      string
	Domain       string
	Root         string
	Flag         int
	CategoryName []string
}

/**
  匹配 host, 要求 host 不包含 port 部分. 返回:
  是否匹配成功, 子域名部分(尾部不包括".")
*/
func (r RootPath) Match(host string) (bool, string) {

	if !strings.HasSuffix(host, r.Domain) {
		return false, ""
	}

	last := len(host) - len(r.Domain)

	switch r.Pattern {
	case "":
		return last == 0, ""
	case ".":
		return last == 0 || host[last-1] == '.', ""
	case "*":
		if last > 1 && host[last-1] == '.' {
			return true, string(host[:last-1])
		}
		return last == 0, ""
	case "?":
		if last > 1 && host[last-1] == '.' {
			return true, string(host[:last-1])
		}
		return last == 0, ""
	}
	return false, ""
}
