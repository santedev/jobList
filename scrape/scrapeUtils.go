package scrape

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unsafe"

	"github.com/PuerkitoBio/goquery"
)

func stringsAllNotVoid(strs ...string) error {
	for _, str := range strs {
		if str == "" {
			return fmt.Errorf("all the strings are void")
		}
	}
	return nil
}

func blackList(company string) error {
	switch company {
	case "BairesDev LLC":
		return fmt.Errorf("this company or whatever it is")
	default:
		return nil
	}
}

func containsNumberAndDollar(s string) bool {
	hasDollar := strings.Contains(s, "$")

	hasNumber := strings.IndexFunc(s, unicode.IsDigit) != -1

	return hasDollar && hasNumber
}

func parseUlIndeed(div *goquery.Selection) (string, error) {
	var bldr strings.Builder
	div.Find("li").Each(func(i int, s *goquery.Selection) {
		s.Contents().Each(func(j int, content *goquery.Selection) {
			if goquery.NodeName(content) == "b" {
				bldr.WriteString(content.Text())
			} else {
				bldr.WriteString(content.Text())
			}
		})
		bldr.WriteString("\n")
	})
	return bldr.String(), nil
}

func parseDivCompu(salary *string, modality *string, div *goquery.Selection) {
	salSpan := div.Find("span.icon.i_salary")
	if salSpan.Length() > 0 {
		*salary = strings.TrimSpace(salSpan.Parent().Text())
	}

	var modSpan *goquery.Selection
	modSpans := []string{
		"span.icon.i_home",
		"span.icon.i_home_office",
		"span.icon.i_office",
	}

	for _, selector := range modSpans {
		modSpan = div.Find(selector)
		if modSpan.Length() > 0 {
			break
		}
	}

	if modSpan.Length() > 0 {
		*modality = strings.TrimSpace(modSpan.Parent().Text())
	}
}

func assignPagesNum(_page string, maxPage string, limit *int, page *int) {
	var err error
	pg, err := strconv.Atoi(_page)
	if err != nil {
		fmt.Println(err.Error())
	}
	mxPg, err := strconv.Atoi(maxPage)

	if mxPg > 0 {
		*limit = mxPg
	}
	if pg > 0 {
		*page = pg
	}

	if err != nil {
		fmt.Println(err.Error())
	}
}

func isChanClosed(ch interface{}) bool {
	if reflect.TypeOf(ch).Kind() != reflect.Chan {
		panic("only channels!")
	}

	// get interface value pointer, from cgo_export
	// typedef struct { void *t; void *v; } GoInterface;
	// then get channel real pointer
	cptr := *(*uintptr)(unsafe.Pointer(
		unsafe.Pointer(uintptr(unsafe.Pointer(&ch)) + unsafe.Sizeof(uint(0))),
	))

	// this function will return true if chan.closed > 0
	// see hchan on https://github.com/golang/go/blob/master/src/runtime/chan.go
	// type hchan struct {
	// qcount   uint           // total data in the queue
	// dataqsiz uint           // size of the circular queue
	// buf      unsafe.Pointer // points to an array of dataqsiz elements
	// elemsize uint16
	// closed   uint32
	// **

	cptr += unsafe.Sizeof(uint(0)) * 2
	cptr += unsafe.Sizeof(unsafe.Pointer(uintptr(0)))
	cptr += unsafe.Sizeof(uint16(0))
	return *(*uint32)(unsafe.Pointer(cptr)) > 0
}

func findStr(listS []string, s string) bool {
	for _, str := range listS {
		str := str
		if str == s {
			return true
		}
	}
	return false
}

func parseDomain(url string) string {
	domain := ""
	switch {
	case strings.Contains(url, "indeed.com"):
		domain = "indeed"
	case strings.Contains(url, "computrabajo.com"):
		domain = "computrabajo"
	case strings.Contains(url, "linkedin.com"):
		domain = "linkedin"
	}
	return domain
}
