package models

type Ilo struct {
	IloIpAddr string `json:"iloIpAddr" gorm:"type:varchar(255)"`
	Port      string `json:"port" gorm:"type:varchar(255)"`
}
