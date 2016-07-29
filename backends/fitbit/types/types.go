package types

type HeartRate struct {
	Activities_heart []struct {
		DateTime string `json:"dateTime"`
		Value    struct {
			CustomHeartRateZones []interface{} `json:"customHeartRateZones"`
			HeartRateZones       []struct {
				CaloriesOut float64 `json:"caloriesOut"`
				Max         int     `json:"max"`
				Min         int     `json:"min"`
				Minutes     int     `json:"minutes"`
				Name        string  `json:"name"`
			} `json:"heartRateZones"`
			RestingHeartRate int `json:"restingHeartRate"`
		} `json:"value"`
	} `json:"activities-heart"`
	Activities_heart_intraday struct {
		Dataset []struct {
			Time  string `json:"time"`
			Value int    `json:"value"`
		} `json:"dataset"`
		DatasetInterval int    `json:"datasetInterval"`
		DatasetType     string `json:"datasetType"`
	} `json:"activities-heart-intraday"`
}

type Sleep struct {
	Sleep []struct {
		AwakeCount      int    `json:"awakeCount"`
		AwakeDuration   int    `json:"awakeDuration"`
		AwakeningsCount int    `json:"awakeningsCount"`
		DateOfSleep     string `json:"dateOfSleep"`
		Duration        int    `json:"duration"`
		Efficiency      int    `json:"efficiency"`
		IsMainSleep     bool   `json:"isMainSleep"`
		LogID           int    `json:"logId"`
		MinuteData      []struct {
			DateTime string `json:"dateTime"`
			Value    string `json:"value"`
		} `json:"minuteData"`
		MinutesAfterWakeup  int    `json:"minutesAfterWakeup"`
		MinutesAsleep       int    `json:"minutesAsleep"`
		MinutesAwake        int    `json:"minutesAwake"`
		MinutesToFallAsleep int    `json:"minutesToFallAsleep"`
		RestlessCount       int    `json:"restlessCount"`
		RestlessDuration    int    `json:"restlessDuration"`
		StartTime           string `json:"startTime"`
		TimeInBed           int    `json:"timeInBed"`
	} `json:"sleep"`
	Summary struct {
		TotalMinutesAsleep int `json:"totalMinutesAsleep"`
		TotalSleepRecords  int `json:"totalSleepRecords"`
		TotalTimeInBed     int `json:"totalTimeInBed"`
	} `json:"summary"`
}

type Activity struct {
	Activities []interface{} `json:"activities"`
	Goals      struct {
		ActiveMinutes int     `json:"activeMinutes"`
		CaloriesOut   int     `json:"caloriesOut"`
		Distance      float64 `json:"distance"`
		Floors        int     `json:"floors"`
		Steps         int     `json:"steps"`
	} `json:"goals"`
	Summary struct {
		ActiveScore      int `json:"activeScore"`
		ActivityCalories int `json:"activityCalories"`
		CaloriesBMR      int `json:"caloriesBMR"`
		CaloriesOut      int `json:"caloriesOut"`
		Distances        []struct {
			Activity string  `json:"activity"`
			Distance float64 `json:"distance"`
		} `json:"distances"`
		Elevation           float64 `json:"elevation"`
		FairlyActiveMinutes int     `json:"fairlyActiveMinutes"`
		Floors              int     `json:"floors"`
		HeartRateZones      []struct {
			CaloriesOut float64 `json:"caloriesOut"`
			Max         int     `json:"max"`
			Min         int     `json:"min"`
			Minutes     int     `json:"minutes"`
			Name        string  `json:"name"`
		} `json:"heartRateZones"`
		LightlyActiveMinutes int `json:"lightlyActiveMinutes"`
		MarginalCalories     int `json:"marginalCalories"`
		RestingHeartRate     int `json:"restingHeartRate"`
		SedentaryMinutes     int `json:"sedentaryMinutes"`
		Steps                int `json:"steps"`
		VeryActiveMinutes    int `json:"veryActiveMinutes"`
	} `json:"summary"`
}

type Devices []Device

type Device struct {
	Battery       string        `json:"battery"`
	DeviceVersion string        `json:"deviceVersion"`
	Features      []interface{} `json:"features"`
	ID            string        `json:"id"`
	LastSyncTime  string        `json:"lastSyncTime"`
	Mac           string        `json:"mac"`
	Type          string        `json:"type"`
}
