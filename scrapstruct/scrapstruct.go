package scrapstruct

type DogForm struct {
	Date 			string `csv:"Date"`
	TrackName		string `csv:"Track"`
	Distance		string `csv:"Distance"`
	SplitTime		string `csv:"Split"`
	Bends			string `csv:"Sections"`
	FinishPosition	string `csv:"-"`
	CompetitorName 	string `csv:"-"`
	Weight			string `csv:"Weight"`
	SectionOneTime	string `csv:"Section 1 Time"`
	FinishTime		string `csv:"Finish Time"`
	
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
	Trap			string
	DogSex			string
	BirthDate		string
	DogSireName		string
	DogDamName		string
	Trainer			string
	Color			string

	CalcRTimeS		string
	SplitTime		string


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
	RTime			string
	RaceDate		string
	RaceTitle		string
	RaceClass		string
	RacePrize		string
	Distance		string
	TrackCondition	string

	TrackDetail		Track
}

type RaceList struct {
	RaceName 		string
	TrackId			string
	Status			string

	Races			[]SubRace
}