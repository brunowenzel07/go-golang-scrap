
package main

import (
	// "encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
	"io/ioutil"
	"github.com/Jeffail/gabs"
	"strings"
)

type TrackResult struct {
	Position		string
	Name 			string
	FinishTime		string
	DogId			string
}

type Track struct {
	Number			string
	Distance		string
	MediaURL		string

	Results			[]TrackResult
}

type SubRace struct {
	RaceId			string
	rTime			string
	raceDate		string

	TrackDetail		Track
}

type RaceList struct {
	RaceName 		string
	TrackId			string

	Races			[]SubRace
}

func sendGetRequestWithURL(url string) []byte {
	// Build the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return nil
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return nil
	}

	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatal(readErr)
		return nil
	}

	return body
}

func GetRaceResult() []RaceList {
	strYesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	getResultUrl := fmt.Sprintf("http://greyhoundbet.racingpost.com/results/blocks.sd?r_date=%s&blocks=header,meetings&_=1", strYesterday);

	getResBody := sendGetRequestWithURL(getResultUrl)
	if getResBody == nil {
		fmt.Println("API call error : " + getResultUrl)
		return nil
	}

	jsonParsed, jsonErr := gabs.ParseJSON(getResBody)
	if jsonErr != nil {
		fmt.Println(jsonErr)
		return nil
	}

	/* --------------- Parse the tracks ------------------- */
	var raceList []RaceList
	tracks, _ := jsonParsed.Path("meetings.tracks").ChildrenMap()

	for _, track := range tracks {
		races, _ := track.Path("races").Children()
		for _, race := range races {
			var raceObj RaceList
			raceObj.RaceName = race.Path("track").Data().(string)
			raceObj.TrackId = race.Path("track_id").Data().(string)

			subRaces, _ := race.Path("races").Children()
			for _, subRace := range subRaces {
				var subRaceObj SubRace
				subRaceObj.RaceId = subRace.Path("raceId").Data().(string)
				subRaceObj.rTime = strings.Split(subRace.Path("rTime").Data().(string), " ")[1]
				subRaceObj.raceDate = strings.Split(subRace.Path("raceDate").Data().(string), " ")[0]

				//Get Race Details 
				subRaceObj = GetRaceDetailResult(subRaceObj, raceObj.TrackId)

				raceObj.Races = append(raceObj.Races, subRaceObj)
			}

			raceList = append(raceList, raceObj);

			// return raceList;
		}
	}

	return raceList;
}

func GetRaceDetailResult(subRaceObj SubRace, trackId string) SubRace{

	paramStr := "&race_id=" + subRaceObj.RaceId
	paramStr += "&track_id=" + trackId
	paramStr += "&r_date=" + subRaceObj.raceDate
	paramStr += "&r_time=" + url.QueryEscape(subRaceObj.rTime)

	url := "http://greyhoundbet.racingpost.com/results/blocks.sd?blocks=meetingHeader,results-meeting-pager,list&_=1" + paramStr
	// fmt.Println("====================================================");
	// fmt.Println("URL : ", url);
	// fmt.Println("Real URL : ", "http://greyhoundbet.racingpost.com/#result-meeting-result/" + paramStr);

	getResBody := sendGetRequestWithURL(url)
	if getResBody == nil {
		fmt.Println("API call error : " + url)
		return subRaceObj
	}

	jsonParsed, jsonErr := gabs.ParseJSON(getResBody)
	if jsonErr != nil {
		fmt.Println(jsonErr)
		return subRaceObj
	}

	/* ------------- Parse the details of track ------------------------ */
	var trackObj Track
	races, _ := jsonParsed.Path("list.track.races").Children()
	
	myRace := races[0]
	for _, r := range races {
		if r.Path("raceId").Data().(string) == subRaceObj.RaceId {
			myRace = r
			break
		}
	}

	raceNumber := strings.Split(myRace.Path("raceTitle").Data().(string), " ")[1]
	raceDistance := myRace.Path("distance").Data().(string)

	mediaURL := ""
	if len(myRace.Path("videoid").Data().(string)) > 0 {
		mediaURL = "http://greyhoundbet.racingpost.com/#result-video/"
		mediaURL += "race_id=" + subRaceObj.RaceId
		mediaURL += "&track_id=" + trackId
		mediaURL += "&r_date=" + subRaceObj.raceDate

		mediaURL += "&video_id=" + myRace.Path("videoid").Data().(string)
		mediaURL += "&clip_id=" + myRace.Path("clipId").Data().(string)
		mediaURL += "&start_sec=" + myRace.Path("startSec").Data().(string)
		mediaURL += "&end_sec=" + myRace.Path("endSec").Data().(string)
	}

	trackObj.Number = raceNumber
	trackObj.Distance = raceDistance
	trackObj.MediaURL = mediaURL

	results, _ := jsonParsed.Path("list.track.results." + subRaceObj.RaceId).Children()
	for _, trackResult := range results {
		var trackResultObj TrackResult

		position := trackResult.Path("position").Data().(string)
		name := trackResult.Path("name").Data().(string)
		time := trackResult.Path("winnersTimeS").Data().(string)
		dogId := trackResult.Path("dogId").Data().(string)

		trackResultObj.Position = position
		trackResultObj.Name = name
		trackResultObj.FinishTime = time
		trackResultObj.DogId = dogId

		trackObj.Results = append(trackObj.Results, trackResultObj)
	}

	subRaceObj.TrackDetail = trackObj

	return subRaceObj;

}

func main() {

	raceResult := GetRaceResult()

	if raceResult == nil {
		fmt.Println("Race List is nil")
		return
	}

	fmt.Println("-------------------raceResult : ", raceResult[0])

}