package helper

type AddressResponse struct {
	Id       uint32 `json:"id"`
	UserID   uint32 `json:"userId"`
	City     string `json:"city"`
	District string `json:"district"`
	State    string `json:"state"`
	Road     string `json:"road"`
}
