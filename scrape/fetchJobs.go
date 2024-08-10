package scrape

import (
	"context"
	"fmt"
	"jobList/handlers/render"
	t "jobList/scrape/types"
	"jobList/store"
	"jobList/views/components"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"github.com/markbates/goth"
)

func GetJobs(w http.ResponseWriter, r *http.Request, sitesObj t.SitesStrct, siteList []string, user goth.User) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("flusher had an error")
	}
	var wg sync.WaitGroup
	ch := make(chan t.JobStrct)
	done := make(chan struct{})
	var page int
	var query string
	countJobs := 0
	var err error

	//array of job's url
	savedOffers, err := store.DB.GetSavedJobOffers(user.UserID)
	if err != nil {
		log.Println(err)
	}

	for _, s := range siteList {
		wg.Add(1)
		go func(site string) {
			defer wg.Done()
			fmt.Println(site, "site")
			page, err = strconv.Atoi(sitesObj[s].Page)
			if err != nil {
				fmt.Println(err.Error())
			}
			query = sitesObj[s].Query
			_ = scrapeJobs(sitesObj[s].Query, sitesObj[s].MaxPage, site, sitesObj[s].Page, ch)
		}(s)
	}
	if page < 1 {
		page = 1
	}

	go func() {
		defer close(done)
		wg.Wait()
	}()

loop:
	for {
		select {
		case <-done:
			close(ch)
			if countJobs <= 10 {
				err = render.Template(w, r, components.NoJobsFound())
			}
			if countJobs > 10 {
				err = render.Template(w, r, components.SendAgain(query, strconv.Itoa(page+1)))
			}
			if err != nil {
				fmt.Println(err.Error())
			}
			flusher.Flush()
			break loop
		case j, ok := <-ch:
			if !ok {
				break loop
			}
			if savedOffers != nil && findStr(savedOffers, j.JobLink) {
				j.Saved = "saved"
			}
			if len(j.Saved) <= 0 && len(user.UserID) > 0 {
				j.Saved = "unsaved"
			}
			if len(j.Saved) <= 0 {
				j.Saved = "login"
			}
			err := render.Template(w, r, components.Jobs(j))
			if err != nil {
				fmt.Println(err.Error())
			}
			countJobs++
			flusher.Flush()
		}
	}
	return nil
}

func scrapeJobs(query string, maxPages string, site string, page string, mainCh chan<- t.JobStrct) error {
	ch := make(chan t.JobStrct)
	doneChan := false
	c := colly.NewCollector(
		colly.UserAgent("Scraperzot/2.0 (+https://website.com/contact)"),
		colly.AllowURLRevisit(),
	)
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*" + site + ".com*",
		Parallelism: 1,
		Delay:       1 * time.Second,
		RandomDelay: 500 * time.Millisecond,
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Request URL: %s failed with error: %v\n", r.Request.URL, err)
		if r.Request.Ctx.GetAny("retry_count") == nil {
			r.Request.Ctx.Put("retry_count", 1)
		}
		retryCountInterface := r.Request.Ctx.GetAny("retry_count")
		var retryCount int
		if retryCountInterface != nil {
			retryCount = retryCountInterface.(int)
		}
		if retryCount >= 3 {
			fmt.Println("Stop retrying after 3 attempts")
			return
		}
		r.Request.Ctx.Put("retry_count", retryCount+1)
		if isChanClosed(mainCh) {
			return
		}
		r.Request.Retry()
	})
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	var err error
	go func() {
		switch site {
		case "computrabajo":
			err = scrapeComputrabajo(c, query, maxPages, site, page, &doneChan, ch)
			if err != nil {
				return
			}
		case "indeed":
			err = scrapeIndeed(c, query, maxPages, site, page, &doneChan, ch)
			if err != nil {
				return
			}
		case "linkedin":
			err = scrapeLinkedin(c, query, site, &doneChan, ch)
			if err != nil {
				return
			}
		default:
			fmt.Println("Unsupported site:", site)
		}
		close(ch)
	}()

loop:
	for {
		select {
		case <-ctx.Done():
			close(ch)
			break loop
		case j, ok := <-ch:
			if err != nil {
				fmt.Println(err.Error())
			}
			if !ok {
				break loop
			}
			mainCh <- j
		}
	}
	doneChan = true
	return nil
}

func scrapeComputrabajo(c *colly.Collector, query string, maxPages string, site string, _page string, doneChan *bool, ch chan<- t.JobStrct) error {
	pageCount, limit, page := 0, 0, 1
	jobs := 0
	var err error
	assignPagesNum(_page, maxPages, &limit, &page)

	url := "https://co.computrabajo.com/trabajo-de-" + strings.Join(strings.Split(query, " "), "-")
	if page > 1 {
		url = url + "?p=" + strconv.Itoa(page)
	}
	c.OnHTML("p.fs24.tc.pAll30.fwB.mbB", func(e *colly.HTMLElement) {
		if len(e.Text) > 0 {
			err = fmt.Errorf("no more jobs or bad query")
		}
	})

	c.OnHTML("article.box_offer", func(e *colly.HTMLElement) {
		title := e.ChildText("h2")
		jobLink := e.ChildAttr("h2 a", "href")
		companyRate := e.ChildText("span.fwB")
		company := e.ChildText("a.fc_base.t_ellipsis")
		date := e.ChildText("p.fs13.fc_aux.mt15")
		place := e.ChildText("span.mr10")
		div := e.DOM.ChildrenFiltered("div.fs13.mt15")
		salary, modality := "", ""
		if div != nil {
			parseDivCompu(&salary, &modality, div)
		}
		if len(jobLink) > 0 {
			jobLink = "https://co.computrabajo.com" + jobLink
		}
		err := stringsAllNotVoid(title, jobLink, company)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		err = blackList(company)
		if err != nil {
			fmt.Println(company)
			return
		}

		job := t.JobStrct{
			Title:       title,
			JobLink:     jobLink,
			CompanyRate: companyRate,
			Company:     company,
			Date:        date,
			Place:       place,
			Salary:      salary,
			Modality:    modality,
			Site:        site,
		}
		if !*doneChan {
			ch <- job
			jobs++
		}
	})
	c.OnHTML("body", func(e *colly.HTMLElement) {
		pageCount++
		u := ""
		if pageCount < limit {
			u = e.ChildAttr("span[title=Siguiente]", "data-path")
		}
		if len(u) > 0 {
			e.Request.Visit(u)
		}
	})
	if err != nil {
		return err
	}
	err = c.Visit(url)
	if err != nil {
		return err
	}
	fmt.Println("done scraping compu. count:", jobs)
	return nil
}

func scrapeIndeed(c *colly.Collector, query string, maxPages string, site string, _page string, doneChan *bool, ch chan<- t.JobStrct) error {
	pageCount, limit, page := 0, 0, 1
	jobs := 0
	var err error
	assignPagesNum(_page, maxPages, &limit, &page)

	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Referer", "https://co.indeed.com/")
	})

	c.OnHTML("h1.css-94bkab.e1tiznh50", func(e *colly.HTMLElement) {
		if len(e.Text) > 0 {
			err = fmt.Errorf("no jobs found or bad query")
		}
	})

	url := "https://co.indeed.com/jobs?q=" + strings.Join(strings.Split(query, " "), "+")
	if page > 1 {
		url = url + "&start=" + strconv.Itoa((page-1)*10)
	}

	c.OnHTML("div.job_seen_beacon", func(e *colly.HTMLElement) {
		title := e.ChildAttr("h2.jobTitle a span", "title")
		company := e.ChildText("div.company_location span.css-63koeb")
		location := e.ChildText("div.company_location div.css-1p0sjhy")
		detailsDiv := e.DOM.ChildrenFiltered("div.css-9446fg")
		jobLink := e.ChildAttr("h2.jobTitle a", "href")
		details := ""
		date := e.ChildText("span.css-qvloho")
		applyVia := e.ChildText("span.ialbl")
		if detailsDiv != nil {
			details, err = parseUlIndeed(detailsDiv)
		}
		if err != nil {
			fmt.Println(err.Error())
		}
		err := stringsAllNotVoid(title, company, jobLink)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if len(jobLink) > 0 {
			jobLink = "https://co.indeed.com/" + jobLink
		}
		job := t.JobStrct{
			Title:    title,
			Company:  company,
			JobLink:  jobLink,
			Location: location,
			Details:  details,
			Date:     date,
			ApplyVia: applyVia,
			Site:     site,
		}
		if !*doneChan {
			ch <- job
			jobs++
		}
	})
	c.OnHTML("html", func(e *colly.HTMLElement) {
		fmt.Println("what up. jobs:", jobs, "hmm:", pageCount, "limit:", limit)
		pageCount++
		u := ""
		if pageCount < limit {
			u = url + "&start=" + strconv.Itoa((pageCount+page-1)*10)
			fmt.Println(u)
			e.Request.Visit(u)
		}
	})
	if err != nil {
		return err
	}
	err = c.Visit(url)
	if err != nil {
		return err
	}
	return nil
}

func scrapeLinkedin(c *colly.Collector, query string, site string, doneChan *bool, ch chan<- t.JobStrct) error {
	jobs := 0
	var err error
	url := "https://www.linkedin.com/jobs/search?keywords=" + strings.Join(strings.Split(query, " "), "+")

	c.OnHTML("section.no-results", func(e *colly.HTMLElement) {
		if len(e.Text) > 0 {
			err = fmt.Errorf("no more results, bad query or no more products")
		}
	})

	c.OnHTML("ul li div.base-card", func(e *colly.HTMLElement) {
		jobLink := e.ChildAttr("a.base-card__full-link", "href")
		title := e.ChildText("h3.base-search-card__title")
		company := e.ChildText("h4.base-search-card__subtitle a")
		location := e.ChildText("span.job-search-card__location")
		date := e.ChildText("time.job-search-card__listdate")
		err := stringsAllNotVoid(jobLink, title, company)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		job := t.JobStrct{
			JobLink:  jobLink,
			Title:    title,
			Company:  company,
			Location: location,
			Date:     date,
			Site:     site,
		}
		if !*doneChan {
			ch <- job
			jobs++
		}
	})
	if err != nil {
		return err
	}
	err = c.Visit(url)
	if err != nil {
		return err
	}
	return nil
}