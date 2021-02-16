package tracker

import (
	"context"
	"log"
	// "time"
	
	"strings"
	"fmt"
	"regexp"
	"strconv"

	"tracker-api-backend-go/api/structs"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/cdp"
)

func Track(item string)[]structs.Product{
	// create chrome instance
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
		
	)
	defer cancel()

	// FIX THIS

	// create a timeout
	// ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	// defer cancel()




	// navigate to a page, wait for an element, click

	var comparePrices []*cdp.Node
	var links []string
	err := chromedp.Run(ctx,
		chromedp.Navigate(`http://google.com/search?q=` + item + `&tbm=shop&tbs=vw:g`),
		// wait for footer element is visible (ie, page is loaded)
		// chromedp.WaitVisible(`//a[@class='a3H7pd']`),
		// find and click "Expand All" link
		chromedp.Nodes(`//a[@class='a3H7pd']`, &comparePrices),

	)
	if err != nil {
		log.Fatal(err)
	}

	selectionNode := comparePrices[0]

	if selectionNode == nil {
		return []structs.Product{structs.Product{Found: false, Company: "", Currency: "", Review: "", Delivery: "", Price: 0.00, TotalPrice: 0.00, URL: ""}} 
	}
	// fmt.Println(selectionNode)

	var nodes []*cdp.Node
	var example string
	if err := chromedp.Run(ctx,
		chromedp.Click(selectionNode.AttributeValue("href")),
		// chromedp.Click(`//a[@class='a3H7pd']`,chromedp.BySearch, chromedp.NodeVisible),

		// retrieve the value of the textarea
		// chromedp.WaitVisible(`#sh-osd__online-sellers-cont`),

		chromedp.Text(`#sh-osd__online-sellers-cont`, &example, chromedp.NodeVisible, chromedp.ByID),
		chromedp.Nodes("//a[@class='sh-osd__seller-link shntl']", &nodes),
	
	); err != nil {
		panic(err)
	}


	// log.Printf("Go's time.After example:\n%s", example)
	s := strings.Split(example, "Visit site")
	regex, _ := regexp.Compile("(?m)^[ \t]*\r?\n")

	for _, n := range nodes {
		links = append(links, "https://www.google.com" + n.AttributeValue("href"))
	}

	for i,_ := range s {
		s[i] = regex.ReplaceAllString(s[i], "")
	}

	items := make([]structs.Product, 0)

	regexPercentagePostive, _ := regexp.Compile("[0-9]+% positive([(]([[0-9]+[,][0-9]+[,]*[0-9]*]*|[0-9]*)[)](\n.*?)*)")
	regexPercentageNegative, _ := regexp.Compile("[0-9]+% negative([(]([[0-9]+[,][0-9]+[,]*[0-9]*]*|[0-9]*)[)](\n.*?)*)")
	regexPricePound, _ := regexp.Compile("£([0-9]+).([0-9]+)")
	regexPriceDollar, _ := regexp.Compile("$([0-9]+).([0-9]+)")
	regexPriceEuro, _ := regexp.Compile("€([0-9]+).([0-9]+)")
	regexFreeDelivery, _ := regexp.Compile("Free delivery")


	// fmt.Println(s[0])
	// fmt.Println(links)

	company, review, delivery, price, totalPrice, URL, currency := "", "", "", 0.00, 0.00, "", ""

			for i, v := range s {
				description := strings.Split(v, "\n")

				tempPrice1 := 0.00
				tempPrice2 := 0.00

				if(i <= len(links) - 1){
					URL = links[i]
				}

				for i, k := range description {

					if i == 0 {
						company = k
					}else if regexPercentagePostive.MatchString(k) || regexPercentageNegative.MatchString(k){
						review = k
					}else if regexPricePound.MatchString(k) || regexPriceEuro.MatchString(k) || regexPriceDollar.MatchString(k){
						if len(k) >= 8{
							f, _ := strconv.ParseFloat(k[2:8], 64)
							if tempPrice1 == 0.00 {
								tempPrice1 = f
							}else {
								tempPrice2 = f
							}
						}else if len(k) < 8 {
							f, _ := strconv.ParseFloat(k[2:], 64)
							if tempPrice1 == 0.00 {
								tempPrice1 = f
							}else {
								tempPrice2 = f
							}
						}
						

						if string([]rune(k)[0]) == "£" {
							currency = "Pound"
						} else if string([]rune(k)[0]) == "$" {
							currency = "Dollar"
						} else if string([]rune(k)[0]) == "€" {
							currency = "Euro"
						}
						
						// if tempPrice1 == 0.00 {
						// 	tempPrice1 = f
						// }else {
						// 	tempPrice2 = f
						// }
					}else if regexFreeDelivery.MatchString(k){
						delivery = k
					}
					
					if i == (len(description) - 1) { 
						if tempPrice1 > tempPrice2 {
							totalPrice, price = tempPrice1, tempPrice2
							r := fmt.Sprintf("%.2f", totalPrice - price)
							delivery = r
						}else if tempPrice1 < tempPrice2 {
							totalPrice, price = tempPrice2, tempPrice1
							r := fmt.Sprintf("%.2f", totalPrice - price)
							delivery = r
						} else if tempPrice1 == 0.00 && tempPrice2 == 0.00{
							break
						}else if tempPrice1 == tempPrice2  {
							totalPrice, price = tempPrice1, tempPrice2
						} 

						items = append(items, structs.Product{Found: true, Company: company, Currency: currency, Review: review, Delivery: delivery, Price: price, TotalPrice: totalPrice, URL: URL})
						company, review, delivery, price, totalPrice, URL, tempPrice1, tempPrice2, currency = "", "", "", 0.00, 0.00, "", 0.00, 0.00, ""
						break
				
					}

				}
			}

		return items



}