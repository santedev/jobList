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
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/markbates/goth"
)

func GetSavedJobs(w http.ResponseWriter, r *http.Request, user goth.User) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("flusher had an error")
	}
	savedOffers, err := store.DB.GetSavedJobOffers(user.UserID)
	if err != nil {
		log.Println(err)
	}
	countJobs := 0
	var wg sync.WaitGroup
	ch := make(chan t.JobStrct)
	done := make(chan struct{})
	for _, offer := range savedOffers {
		wg.Add(1)
		go func(offer string) {
			defer wg.Done()
			err = scrapeSavedJobs(offer, user.UserID, ch)
		}(offer)
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
			if countJobs == 0 {
				err = render.Template(w, r, components.NoJobsFound())
			}
			flusher.Flush()
			break loop
		case j, ok := <-ch:
			if !ok {
				break loop
			}
			if err != nil {
				log.Println(err)
			}
			err := render.Template(w, r, components.Jobs(j))
			if err != nil {
				log.Println(err)
			}
			countJobs++
			flusher.Flush()
		}
	}

	return nil
}

func scrapeSavedJobs(offerLink string, uid string, mainCh chan<- t.JobStrct) error {
	var err error
	ch := make(chan t.JobStrct)
	c := colly.NewCollector(
		colly.UserAgent("Scraperzot/2.0 (+https://website.com/contact)"),
		colly.AllowURLRevisit(),
	)

	site := parseDomain(offerLink)
	if len(site) <= 0 {
		return fmt.Errorf("the link has no domain")
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

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

	go func() {
		switch site {
		case "indeed":
			err = scrapeOfferIndeed(c, site, offerLink, ch)
		case "computrabajo":
			err = scrapeOfferComputrabajo(c, site, offerLink, ch)
		case "linkedin":
			err = scrapeOfferLinkedin(c, site, offerLink, ch)
		default:
			err = fmt.Errorf("no site with that name; tried to scrape wrong site")
		}
		if err != nil {
			return
		}
		close(ch)
	}()

loop:
	for {
		select {
		case <-timeoutCtx.Done():
			close(ch)
			break loop
		case j, ok := <-ch:
			if !ok {
				break loop
			}
			if err != nil {
				err = store.DB.UnsaveJobOffer(uid, offerLink)
				break loop
			}

			j.Saved = "saved"
			j.JobLink = offerLink
			mainCh <- j
		}
	}
	return err
}

func scrapeOfferComputrabajo(c *colly.Collector, site string, url string, ch chan<- t.JobStrct) error {
	c.OnHTML("main.detail_fs", func(e *colly.HTMLElement) {
		title := e.ChildText("h1.fwB.fs24")
		location := e.ChildText("p.fs16")
		salary := e.DOM.Find("div.mbB span.tag.base.mb10").Text()
		if !containsNumberAndDollar(salary) {
			salary = ""
		}
		description := e.DOM.Find("p.mbB").Text()

		ul := e.DOM.Find("ul.disc")
		requirements := ul.Text()
		var bulletPoints string
		ul.Find("li").Each(func(i int, s *goquery.Selection) {
			bulletPoints += "- " + s.Text() + "\n"
		})

		datePosted := e.ChildText("p.fc_aux.fs13")

		companyName := e.ChildText("div.info_company a.js-o-link")
		rating := e.ChildText("div.info_company div.fs16 span.star + span")

		job := t.JobStrct{
			Title:       title,
			Company:     companyName,
			Location:    location,
			Details:     description + "\n" + requirements + "\n" + bulletPoints,
			Salary:      salary,
			Date:        datePosted,
			CompanyRate: rating,
			Site:        site,
		}

		ch <- job
	})

	c.Visit(url)
	return nil
}

func scrapeOfferIndeed(c *colly.Collector, site string, url string, ch chan<- t.JobStrct) error {
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Referer", "https://co.indeed.com/")
	})
	c.OnHTML("div.jobsearch-JobComponent.css-u4y1in", func(e *colly.HTMLElement) {
		title := e.ChildText("h1.jobsearch-JobInfoHeader-title span")
		companyRate := e.ChildText("span.css-1b6omqv.esbq1260")
		companyName := e.ChildText("a.css-1ioi40n.e19afand0")
		companyLocation := e.ChildText("div[data-testid=jobsearch-JobInfoHeader-companyLocation] span")
		description := e.DOM.Find("#jobDescriptionText").Text()
		salary := e.ChildText("div.js-match-insights-provider-tvvxwd.ecydgvn1")
		err := stringsAllNotVoid(title, companyName)
		if err != nil {
			return
		}
		job := t.JobStrct{
			Title:       title,
			Company:     companyName,
			CompanyRate: companyRate,
			Location:    companyLocation,
			Details:     description,
			Salary:      salary,
			Site:        site,
		}
		ch <- job
	})
	c.Visit(url)
	return nil
}

func scrapeOfferLinkedin(c *colly.Collector, site string, url string, ch chan<- t.JobStrct) error {
	c.OnHTML("div.mt4[role=main]", func(e *colly.HTMLElement) {
		title := e.ChildText("div.job-details-jobs-unified-top-card__job-title h1")
		companyName := e.ChildText("div.job-details-jobs-unified-top-card__company-name a")
		companyLocation := e.ChildText("div.job-details-jobs-unified-top-card__primary-description-container span.tvm__text")
		description := e.DOM.Find("article.jobs-description__container").Text()
		salary := e.ChildText("#SALARY div.artdeco-card.mt4 p.t-16")

		err := stringsAllNotVoid(title, companyName)
		if err != nil {
			return
		}

		job := t.JobStrct{
			Title:    title,
			Company:  companyName,
			Location: companyLocation,
			Details:  description,
			Salary:   salary,
			Site:     site,
		}
		ch <- job
	})
	c.Visit(url)
	return nil
}