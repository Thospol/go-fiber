package user

type getUserRequest struct {
	Id uint `form:"id" json:"id" path:"id" query:"id" xml:"id"`
}
