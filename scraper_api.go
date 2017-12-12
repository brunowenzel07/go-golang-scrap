
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
	"strconv"

	Structs "./scrapstruct"
)

type Venue struct {
	venue_id		string
	provider_id		string
	venue_type		string
	venue_name		string

	itsp_codes		[]string
	mapping_itsp_code string
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

func GetRaceResult() []Structs.RaceList {
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
	var raceList []Structs.RaceList
	tracks, _ := jsonParsed.Path("meetings.tracks").ChildrenMap()

	for _, track := range tracks {
		races, _ := track.Path("races").Children()
		for _, race := range races {
			if race.Path("meeting_abandoned").Data().(float64) == 1 {
				continue;
			}

			var raceObj Structs.RaceList
			raceObj.RaceName = race.Path("track").Data().(string)
			raceObj.TrackId = race.Path("track_id").Data().(string)
			raceObj.Status = "CLOSED"

			subRaces, _ := race.Path("races").Children()
			for _, subRace := range subRaces {
				var subRaceObj Structs.SubRace
				subRaceObj.RaceId = subRace.Path("raceId").Data().(string)
				subRaceObj.RaceTitle = subRace.Path("raceTitle").Data().(string)
				subRaceObj.RTime = strings.Split(subRace.Path("rTime").Data().(string), " ")[1]
				subRaceObj.RaceDate = strings.Split(subRace.Path("raceDate").Data().(string), " ")[0]
				subRaceObj.RaceClass = subRace.Path("raceGrade").Data().(string)
				subRaceObj.Distance = subRace.Path("distance").Data().(string)
				subRaceObj.TrackCondition = subRace.Path("raceType").Data().(string)

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

func GetRaceDetailResult(subRaceObj Structs.SubRace, trackId string) Structs.SubRace{

	paramStr := "&race_id=" + subRaceObj.RaceId
	paramStr += "&track_id=" + trackId
	paramStr += "&r_date=" + subRaceObj.RaceDate
	paramStr += "&r_time=" + url.QueryEscape(subRaceObj.RTime)

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
	var trackObj Structs.Track
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
		mediaURL += "&r_date=" + subRaceObj.RaceDate

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
		var trackResultObj Structs.TrackResult

		position := trackResult.Path("position").Data().(string)
		name := trackResult.Path("name").Data().(string)
		time := trackResult.Path("winnersTimeS").Data().(string)
		dogId := trackResult.Path("dogId").Data().(string)

		trackResultObj.Position 	= position
		trackResultObj.Name 		= name
		trackResultObj.FinishTime 	= time
		trackResultObj.DogId 		= dogId
		trackResultObj.Trap 		= trackResult.Path("trap").Data().(string)
		trackResultObj.DogSex 		= trackResult.Path("dogSex").Data().(string)
		trackResultObj.BirthDate 	= strings.Split(trackResult.Path("dogDateOfBirth").Data().(string), " ")[0]
		trackResultObj.DogSireName 	= trackResult.Path("dogSire").Data().(string)
		trackResultObj.DogDamName 	= trackResult.Path("dogDam").Data().(string)
		trackResultObj.Trainer 		= trackResult.Path("trainer").Data().(string)
		trackResultObj.Color 		= trackResult.Path("dogColor").Data().(string)

		//Temporary comment for further details
		// dogObj := GetDogDetail( subRaceObj.RaceId, trackId, dogId, subRaceObj.RaceDate, subRaceObj.RTime)
		// trackResultObj.Dog = dogObj

		trackObj.Results = append(trackObj.Results, trackResultObj)
	}

	subRaceObj.TrackDetail = trackObj

	return subRaceObj;

}

func GetDogDetail(raceId string, trackId string, dogId string, r_date string, r_time string) Structs.Dog {

	var dogObj Structs.Dog

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

		var dogFormObj Structs.DogForm
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

	//Assume provider_id is given (= 41)
	//Assume venue_type is given (= GREYHOUND)
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
		venueObj.provider_id = "41"  //Assume provider_id is given(41)
		venueObj.venue_id = fmt.Sprintf("%g", venue.Path("venue_id").Data().(float64));
		venueObj.venue_type = venue.Path("venue_type").Data().(string)
		venueObj.venue_name = venue.Path("name").Data().(string)

		if venue.Path("mapping.itsp_code").Data() != nil {
			venueObj.mapping_itsp_code = venue.Path("mapping.itsp_code").Data().(string)	
		}

		itspCodes, _ := venue.Path("itsp_codes").ChildrenMap()
		for _, code := range itspCodes {
			venueObj.itsp_codes = append(venueObj.itsp_codes, code.Path("itsp_code").Data().(string))	
		}

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

func PostPayloadForData(token string, payload string) bool{
	//Assume provider_code is given (= racingpost)
	url := "https://api-test.betia.co/tmp/providers/racingpost/meetings"
	
	//Build the request
	var jsonData = payload;
	var jsonStr = []byte(jsonData)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return false
	}

	//Set the headers
	bearTokenStr := "Bearer " + token;
	req.Header.Set("Authorization", bearTokenStr)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return false
	}

	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatal(readErr)
		return false
	}

	fmt.Println("Response Status : " + resp.Status);
	fmt.Println("Response Body : " + string(body));
	if resp.Status == "202 Accepted" { //Success
		return true;
	}
	
	return false
}

func PostAllPayloadsWithRaceResult(raceResult []Structs.RaceList) {
	//Get a token befor posting
	token := GetToken();
	if len(token) == 0 {
		fmt.Printf("Token fetch is Failed!!!--------")
		fmt.Println()
		return
	}

	//Post all race results to the APIs
	for _, raceObj := range raceResult {
		//Initialize JSON Obj
		jsonObj := gabs.New()
		jsonObj.SetP(41, "provider_id") //Asume provider_id is 41
		jsonObj.SetP("scapers-go-raymond", "source")

		//Get a Venue detail from API
		venueName := raceObj.RaceName
		venue := GetVenueDetail(venueName)
		if len(venue.venue_id) == 0 {
			fmt.Printf("***The venue '%s' is not existing on the server***", venueName)
			fmt.Println();
			continue
		}

		jsonObj.SetP(raceObj.Status, "status")
		jsonObj.SetP(venue.venue_name, "venue_name")
		jsonObj.SetP(venue.venue_type, "venue_type")
		jsonObj.SetP(time.Now().AddDate(0, 0, -1).Format("2006-01-02"), "meeting_date")
		venue_id, _ := strconv.Atoi(venue.venue_id)
		jsonObj.SetP(venue_id, "venue_id")
		if venue.mapping_itsp_code == "" {
			jsonObj.SetP(nil, "itsp_code")
		} else {
			jsonObj.SetP(venue.mapping_itsp_code, "itsp_code")
		}
		jsonObj.ArrayP("venue_itsp_codes")
		for _, itsp_code := range venue.itsp_codes {
			jsonObj.ArrayAppend(itsp_code, "venue_itsp_codes");
		}

		jsonObj.ArrayP("events")
		for raceIndex, subRaceObj := range raceObj.Races {
			//Initialize Event Object
			eventObj := make(map[string]interface{})
			eventObj["number"] = raceIndex + 1
			eventObj["status"] = "CLOSED"
			eventObj["name"] = subRaceObj.RaceTitle

			layout 	:= "2006-01-02 15:04"
			timeStr := subRaceObj.RaceDate + " " + subRaceObj.RTime
			t, _ := time.Parse(layout, timeStr)
			t = t.Add(time.Hour * time.Duration(-2))
			eventObj["start_time"] = t.Unix()

			//Race Data
			raceDataObj := make(map[string]interface{})
			raceDataObj["race_class"] 		= subRaceObj.RaceClass
			raceDataObj["distance_unit"] 	= "yards"

			distanceNumber, _ := strconv.Atoi(subRaceObj.Distance)
			raceDataObj["distance"] 		= distanceNumber
			if subRaceObj.TrackCondition == "F" {
				raceDataObj["track_condition"] 	= "Flat"
			} else {
				raceDataObj["track_condition"] 	= subRaceObj.TrackCondition
			}

			eventObj["race_data"] = raceDataObj

			//Competitors
			var competitorsArray []map[string]interface{}
			for _, trackDetailObj := range subRaceObj.TrackDetail.Results {
				competitor := make(map[string]interface{})
				competitor["name"] 			= trackDetailObj.Name
				competitor["birth_date"] 	= trackDetailObj.BirthDate
				competitor["colour"] 		= trackDetailObj.Color

				if trackDetailObj.DogSex == "D" {
					competitor["sex"] 		= "DOG"
				} else {
					competitor["sex"] 		= "BITCH"
				}

					sireObj := make(map[string]string)
					sireObj["name"] = trackDetailObj.DogSireName
				competitor["sire"] 			= sireObj

					damObj := make(map[string]string)
					damObj["name"] = trackDetailObj.DogDamName
				competitor["dam"] 			= damObj

					racingObj := make(map[string]interface{})
					positionNumber, _ := strconv.Atoi(trackDetailObj.Position)
					racingObj["number"] = positionNumber
					trapNumber, _ := strconv.Atoi(trackDetailObj.Trap)
					racingObj["barrier"] = trapNumber
					racingObj["set_last_known_weight"] = true

					trainerObj := make(map[string]string)
					trainerObj["name"] = trackDetailObj.Trainer
					trainerObj["jurisdiction"] = "Ireland"
					racingObj["trainer"] = trainerObj
				competitor["race_data"] 	= racingObj

					metaDataObj := make(map[string]interface{})
					var racingpostObj []map[string]interface{}
						racingpostItem := make(map[string]interface{})
						racingpostItem["key"] = "runner_id"
						racingpostItem["value"] = trackDetailObj.DogId

					racingpostObj = append(racingpostObj, racingpostItem)
					metaDataObj["racingpost"] = racingpostObj

				competitor["metadata"] 	= metaDataObj

				competitorsArray = append(competitorsArray, competitor)
			}
			eventObj["competitors"] = competitorsArray

			jsonObj.ArrayAppend(eventObj, "events");
		}


		// fmt.Println("-------------------Payload Result : ", jsonObj.String())
		
		//Get the payload as a strong from JSON object
		payload := jsonObj.String();

		//Post the payload
		success := PostPayloadForData(token, payload)
		if success == true {
			fmt.Println("Post SUCCESS!!!!!--- Venue : ", venueName)
		} else {
			fmt.Println("Post FAILURE!!!!!--- Venue : ", venueName)
		}


	}
}

func main() {
	//Get Race result using a scrapper
	raceResult := GetRaceResult()
	if raceResult == nil {
		fmt.Println("Race List is nil")
		return
	}

	fmt.Println("RaceResults are all fetched!!!-------")
	// fmt.Println("-------------------raceResult : ", raceResult[0])
	PostAllPayloadsWithRaceResult(raceResult);
}