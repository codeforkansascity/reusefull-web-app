package models

type Charity struct {
	Id                      int      `json:"id"`
	Name                    string   `json:"name"`
	ContactName             string   `json:"contactName"`
	Phone                   string   `json:"phone"`
	Website                 string   `json:"website"`
	Email                   string   `json:"email"`
	Faith                   *bool    `json:"faith"`
	Pickup                  bool     `json:"pickup"`
	Dropoff                 bool     `json:"dropoff"`
	Resell                  *bool    `json:"resell"`
	NewItems                *bool    `json:"newItems"`
	AmazonWishlist          string   `json:"amazon"`
	GoodItems               bool     `json:"goodItems"`
	CashDonationLink        string   `json:"cashDonate"`
	VolunteerSignup         string   `json:"volunteer"`
	Address                 string   `json:"address"`
	City                    string   `json:"city"`
	State                   string   `json:"state"`
	ZipCode                 string   `json:"zip"`
	LogoURL                 string   `json:"logoURL"`
	Logo                    string   `json:"logo"`
	Lat                     float64  `json:"lat"`
	Lng                     float64  `json:"lng"`
	Mission                 string   `json:"mission"`
	Description             string   `json:"description"`
	ItemTypes               []string `json:"itemTypes"`
	ItemTypeDescriptions    []string `json:"itemTypeDescriptions"`
	CharityTypes            []string `json:"charityTypes"`
	CharityTypeDescriptions []string `json:"charityTypeDescriptions"`
	CharityTypeOther        string   `json:"other"`
	Budget                  string   `json:"budget"`
	TaxID                   string   `json:"taxID"`
	UserID                  string   `json:"userID"`
	Approved                bool     `json:"approved"`
	EmailVerified           bool     `json:"emailVerified"`
	Paused                  bool     `json:"paused"`
	Facebook                string   `json:"facebook"`
	Twitter                 string   `json:"twitter"`
	Instagram               string   `json:"instagram"`
	YouTube                 string   `json:"youtube"`
	Snapchat                string   `json:"snapchat"`
	TikTok                  string   `json:"tiktok"`
}

type CharityType struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}