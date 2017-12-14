package scrapstruct

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