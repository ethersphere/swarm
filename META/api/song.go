package api

import (

)

type Song struct {
	api *Api
}

func NewSong(api *Api) *Song {
	return &Song{api}
}

func (self *Song) Echo(s string) (string,error) {
	return s, nil
}
