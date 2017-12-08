
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
	"bytes"
)

type DogForm struct {
	Date 			string
	TrackName		string
	Distance		string
	Bends			string
	FinishPosition	string
	CompetitorName 	string
	Weight			string
	FinishTime		string
}

type Dog struct {
	Name 			string
	SireName		string
	DamName			string

	Forms			[]DogForm
}

type TrackResult struct {
	Position		string
	Name 			string
	FinishTime		string
	DogId			string

	Dog				Dog
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



type Venue struct {
	venue_id		string
	provider_id		string
	venue_type		string
	venue_name		string
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

		dogObj := GetDogDetail( subRaceObj.RaceId, trackId, dogId, subRaceObj.raceDate, subRaceObj.rTime)
		trackResultObj.Dog = dogObj

		trackObj.Results = append(trackObj.Results, trackResultObj)
	}

	subRaceObj.TrackDetail = trackObj

	return subRaceObj;

}

func GetDogDetail(raceId string, trackId string, dogId string, r_date string, r_time string) Dog {

	var dogObj Dog

	paramStr := "&race_id=" + raceId
	paramStr += "&track_id=" + trackId
	paramStr += "&dog_id=" + dogId
	paramStr += "&r_date=" + r_date
	paramStr += "&r_time=" + url.QueryEscape(r_time)

	url := "http://greyhoundbet.racingpost.com/results/blocks.sd?blocks=results-dog-details&_=1" + paramStr
	// fmt.Println("====================================================");
	// fmt.Println("URL : ", url);
	// fmt.Println("Real URL : ", "http://greyhoundbet.racingpost.com/#results-dog/" + paramStr);

	getResBody := sendGetRequestWithURL(url)
	if getResBody == nil {
		fmt.Println("API call error : " + url)
		return dogObj
	}

	jsonParsed, jsonErr := gabs.ParseJSON(getResBody)
	if jsonErr != nil {
		fmt.Println(jsonErr)
		return dogObj
	}

	/* ----------------------- Parse Dog Info -------------------------- */

	dogInfo 		:= jsonParsed.Path("results-dog-details.dogInfo")
	dogObj.Name 	= dogInfo.Path("dogName").Data().(string)
	dogObj.SireName = dogInfo.Path("sireName").Data().(string)
	dogObj.DamName 	= dogInfo.Path("damName").Data().(string)
	
	formsInfo, _ := jsonParsed.Path("results-dog-details.forms").Children()
	for _, form := range formsInfo {

		var dogFormObj DogForm
		dogFormObj.Date 			= form.Path("rFormDatetime").Data().(string)
		dogFormObj.TrackName 	 	= form.Path("trackShortName").Data().(string)
		dogFormObj.Distance			= form.Path("distMetre").Data().(string)
		dogFormObj.Bends			= form.Path("bndPos").Data().(string)
		dogFormObj.FinishPosition	= form.Path("rOutcomeDesc").Data().(string)
		dogFormObj.CompetitorName 	= form.Path("otherDogName").Data().(string)
		dogFormObj.Weight			= form.Path("weight").Data().(string)
		dogFormObj.FinishTime		= form.Path("winnersTimeS").Data().(string)

		dogObj.Forms = append(dogObj.Forms, dogFormObj)
	}

	return dogObj
}

func GetVenueDetail(venueName string) Venue{
	var venueObj Venue

	//Assume provider_id is given(41)
	//Assume venue_type is given(GREYHOUND)
	getResultUrl := fmt.Sprintf("https://staging.dw.xtradeiom.com/api/venues/search?venue_name=%s&venue_type=GREYHOUND&provider_id=41", url.QueryEscape(venueName));
	
	getResBody := sendGetRequestWithURL(getResultUrl)
	if getResBody == nil {
		fmt.Println("API call error : " + getResultUrl)
		return venueObj
	}

	jsonParsed, jsonErr := gabs.ParseJSON(getResBody)
	if jsonErr != nil {
		fmt.Println(jsonErr)
		return venueObj
	}

	venues, _ := jsonParsed.Path("data.venues").ChildrenMap()
	for _, venue := range venues {
		venueObj.venue_id = fmt.Sprintf("%g", venue.Path("venue_id").Data().(float64));
		venueObj.venue_type = venue.Path("venue_type").Data().(string)
		venueObj.venue_name = venue.Path("name").Data().(string)
		venueObj.provider_id = "41"  //Assume provider_id is given(41)

		return venueObj;
	}

	return venueObj;
}

func GetToken() string {
	url := "https://auth.betia.co/auth"

	//Build the request
	var jsonData = `{"email":"scraper@betia.co","password":"L9x?E63h4H=6"}`;
	var jsonStr = []byte(jsonData)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return ""
	}

	//Set the headers
    req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return ""
	}

	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatal(readErr)
		return ""
	}

	if resp.Status == "200 OK" { //Success
		return string(body);
	}
	
	return ""
}

func PostToken() string {
	
}

func main() {

	//------------Get a token------------- //
	token := GetToken();
	if len(token) == 0 {
		fmt.Printf("Token fetch is Failed.")
		return;
	}
	fmt.Println("-------------------Token : ", token)

	//------------Get a Venue detail------------- //
	// venueName := "Belle Vue1"
	// venue := GetVenueDetail(venueName)
	// if len(venue.venue_id) == 0 {
	// 	fmt.Printf("The venue '%s' is not existing on the server", venueName)
	// 	return;
	// }
	// fmt.Println("-------------------Venue Result : ", venue)

	//------------Get Race result using a scrapper ------------- //
	// raceResult := GetRaceResult()

	// if raceResult == nil {
	// 	fmt.Println("Race List is nil")
	// 	return
	// }

	// fmt.Println("-------------------raceResult : ", raceResult[0])

}