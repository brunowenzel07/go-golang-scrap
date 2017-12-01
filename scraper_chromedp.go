package main

import (
	"context"
	"log"
	"time"
	"strings"

	cdp "github.com/knq/chromedp"
	// cdptypes "github.com/knq/chromedp/cdp"
)

func main() {
	var err error

	// create context
	ctxt, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create chrome instance
	// c, err := cdp.New(ctxt, cdp.WithLog(log.Printf))
	c, err := cdp.New(ctxt)
	if err != nil {
		log.Fatal(err)
	}

	// Run task to get the RaceList
	var resRaceList [][]string
	err = c.Run(ctxt, getTodayResults(&resRaceList))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("====================RaceList=============================\n%#v", resRaceList)

	//Run task to get the result dogs
	for i := 0; i < len(resRaceList); i++ {

		for j := 0; j < (len(resRaceList[i]) - 1) / 2; j++ {
			indexDog := (j + 1) * 2;
			raceDetailLink := resRaceList[i][indexDog];

			var resResultDogs [][]string
			var resResultTitle string
			var resWatchURL string
			err = c.Run(ctxt, getResultsDogs(raceDetailLink, &resResultTitle, &resWatchURL, &resResultDogs))
			if err != nil {
				log.Fatal(err)
			}

			var eventNumber string = strings.Split(resResultTitle, " ")[1]
			var eventDistance string = strings.Split(resResultTitle, " ")[4]
			log.Printf("=================================================");
			log.Printf("raceDetailLink: %s", raceDetailLink)
			// log.Printf("Title: %s", resResultTitle)
			log.Printf("eventNumber: %s", eventNumber)
			log.Printf("eventDistance: %s", eventDistance)
			log.Printf("URL: %s", resWatchURL)
			log.Printf("Dogs: %s", resResultDogs)
			

			//Run task to get the runner info
			for k := 0; k < len(resResultDogs); k++ {
				runnerLink := resResultDogs[k][0];

				var resLifeTimeInfo [][]string
				var resName string
				var resSireName string
				var resDamName string
				err = c.Run(ctxt, getResultsRunner(runnerLink, &resName, &resSireName, &resDamName, &resLifeTimeInfo))
				if err != nil {
					log.Fatal(err)
				}

				log.Printf("=================================================");
				log.Printf("Name: %s", resName)
				log.Printf("resSireName: %s", resSireName)
				log.Printf("resDamName: %s", resDamName)
				log.Printf("resLifeTimeInfo: %v", resLifeTimeInfo)

			}
		}
	}



	// shutdown chrome
	err = c.Shutdown(ctxt)
	if err != nil {
		log.Fatal(err)
	}

	// wait for chrome to finish
	err = c.Wait()
	if err != nil {
		log.Fatal(err)
	}

	
	
}

func jsGetRaceList(sel string) (js string) {
	const funcJS = `function getText(sel) {
				var text = [];
				var elements = document.body.querySelectorAll(sel);

				for(var i = 0; i < elements.length; i++) {
					var raceInfo = [];
					var current = elements[i];

					var raceName = current.querySelector('.results-race-name h4').innerHTML;
					raceInfo.push(raceName);
					var raceList = [];

					timeElements = current.querySelectorAll('.results-race-list-wrapper .results-race-list-row a')
					for(var k = 0; k < timeElements.length; k++) {
						timeElement = timeElements[k];

						var timeStr = timeElement.innerHTML;
						var timeLink = timeElement.href;

						raceInfo.push(timeStr);
						raceInfo.push(timeLink);
					}

					text.push(raceInfo);
					// if(current.children.length === 0 && current.textContent.replace(/ |\n/g,'') !== '') {
					// Check the element has no children && that it is not empty
					// 	text.push(current.textContent + ',');
					// }
				}
				return text
			 };`

	invokeFuncJS := `var a = getText('` + sel + `'); a;`
	return strings.Join([]string{funcJS, invokeFuncJS}, " ")
}

func getTodayResults(res *[][]string) cdp.Tasks {
	today := time.Now()
	yesterday := today.AddDate(0, 0, -1)
	strYesterday := yesterday.Format("2006-01-02")

	textSel := ".raceList li"
	jsText := jsGetRaceList(textSel)

	// fmt.Println("jsText: " + jsText);

	return cdp.Tasks{
		cdp.Navigate("http://greyhoundbet.racingpost.com/#results-list/r_date=" + strYesterday),
		cdp.WaitVisible("latestResults", cdp.ByID),
		cdp.Evaluate(jsText, res),
	}
}

//------------------------------ Result dogs ------------------------------- //
func jsGetResultDogs(sel string) (js string) {
	const funcJS = `function getText(sel) {
				var text = [];
				var elements = document.body.querySelectorAll(sel);

				for(var i = 0; i < elements.length; i++) {
					var current = elements[i];
					var refLink = current.href;

					var result = current.querySelector('.result');
					var finishPosition = result.querySelector('.place').innerHTML.split('<')[0].trim();
					var competitorName = result.querySelector('.info .holder .result-dog-name-details .name').innerHTML.trim();
					var finishTime = result.querySelector('.info .holder .dog-cols .col1').innerHTML.replace(/\&nbsp;/g, '').trim();

					var dogInfo = [];
					dogInfo.push(refLink);
					dogInfo.push(finishPosition);
					dogInfo.push(competitorName);
					dogInfo.push(finishTime);

					text.push(dogInfo);
				}
				return text
			 };`

	invokeFuncJS := `var a = getText('` + sel + `'); a;`
	return strings.Join([]string{funcJS, invokeFuncJS}, " ")
}

func getResultsDogs(urlLink string, resTitle *string, resMediaURL *string, res *[][]string) cdp.Tasks {
	textSel := ".meetingResultsList .container .details"
	jsText := jsGetResultDogs(textSel)

	// fmt.Println("jsText: " + jsText);

	return cdp.Tasks{
		cdp.Navigate(urlLink),
		cdp.WaitNotPresent("pageLoading", cdp.ByID),
		cdp.WaitVisible("circle-race-title", cdp.ByID),
		cdp.Text("circle-race-title", resTitle),
		cdp.EvaluateAsDevTools("document.querySelector('.raceTitle .buttonsBox a') ? document.querySelector('.raceTitle .buttonsBox a').href : ''", resMediaURL),
		cdp.Evaluate(jsText, res),
	}
}



//------------------------------ Result runner ------------------------------- //
func jsGetResultRunner(sel string) (js string) {
	const funcJS = `function getText(sel) {
				var text = [];
				var elements = document.body.querySelectorAll(sel);

				//Skip the first row which is a header
				for(var i = 1; i < elements.length; i++) {
					var current = elements[i];

					var tdElements = current.querySelectorAll('td');
					var date = tdElements[0].contentText;
					var venueName = tdElements[1].innerHTML;
					var distance = tdElements[2].innerHTML;
					var sectionalData = tdElements[5].innerHTML;
					var finishingPos = tdElements[6].innerHTML;
					var weight = tdElements[12].innerHTML;
					var finishingTime = tdElements[15].innerHTML;
					var mediaURL = tdElements[0].querySelector('a') ? tdElements[0].querySelector('a').href : '';
					var competitorName = tdElements[8].textContent;

					var runnerInfo = [];
					runnerInfo.push(date);
					runnerInfo.push(venueName);
					runnerInfo.push(distance);
					runnerInfo.push(sectionalData);
					runnerInfo.push(finishingPos);
					runnerInfo.push(weight);
					runnerInfo.push(finishingTime);
					runnerInfo.push(mediaURL);
					runnerInfo.push(competitorName);
					

					text.push(runnerInfo);
				}
				return text
			 };`

	invokeFuncJS := `var a = getText('` + sel + `'); a;`
	return strings.Join([]string{funcJS, invokeFuncJS}, " ")
}

func getResultsRunner(urlLink string, resName *string, resSireName *string, resDamName *string, res *[][]string) cdp.Tasks {
	textSel := "#sortableTable tbody tr"
	jsText := jsGetResultRunner(textSel)

	// log.Println("jsText: " + jsText);

	return cdp.Tasks{
		cdp.Navigate(urlLink),
		cdp.WaitNotPresent("pageLoading", cdp.ByID),
		cdp.EvaluateAsDevTools("document.querySelector('.results-dog-details .ghName').textContent.trim();", resName),
		cdp.EvaluateAsDevTools("document.querySelector('.results-dog-details .runnerBlock .pedigree tbody tr td strong').innerHTML.trim();", resSireName),
		cdp.EvaluateAsDevTools("document.querySelector('.results-dog-details .runnerBlock .pedigree tbody tr td:last-child strong').innerHTML.trim();", resDamName),
		cdp.Evaluate(jsText, res),
	}
}