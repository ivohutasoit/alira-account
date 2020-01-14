package controller

import (
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/ivohutasoit/alira-account/constant"
	"github.com/ivohutasoit/alira-account/service"
)

var validate *validator.Validate

func IdentityHandler(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		c.HTML(http.StatusOK, constant.IdentityPage, nil)
		return
	}

	type Request struct {
		Document       string    `form:"document" json:"document" xml:"document" binding:"required"`
		NationID       string    `form:"nation_id" json:"nation_id" xml:"nation_id" binding:"required"`
		Fullname       string    `form:"fullname" json:"fullname" xml:"fullname" binding:"required"`
		BirthPlace     string    `form:"birth_place" json:"birth_place" xml:"birth_place"`
		BirthDate      string    `form:"birth_date" json:"birth_date" xml:"birth_date"`
		Address        string    `form:"address" json:"address" xml:"address"`
		City           string    `form:"city" json:"city" xml:"city"`
		State          string    `form:"state" json:"state" xml:"state"`
		Province       string    `form:"province" json:"province" xml:"province"`
		Country        string    `form:"country" json:"country" xml:"country" binding:"required"`
		PostalCode     string    `form:"postal_code" json:"postal_code" xml:"postal_code"`
		BloodType      string    `form:"blood_type" json:"blood_type" xml:"blood_type"`
		Religion       string    `form:"religion" json:"religion" xml:"religion"`
		MarriedStatus  string    `form:"married_status" json:"married_status" xml:"married_status"`
		Type           string    `form:"type" json:"type" json:"type"`
		Nationality    string    `form:"nationality" json:"nationality" xml:"nationality"`
		IssueDate      time.Time `form:"issued_date" json:"issued_date" xml:"issued_date" time_format:"2006-01-02"`
		ExpiryDate     time.Time `form:"expiry_date" json:"expiry_date" xml:"expiry_date" time_format:"2006-01-02"`
		RegistrationNo string    `form:"reg_no" json:"reg_no" xml:"reg_no"`
		IssuedOffice   string    `form:"issued_office" json:"issued_office" xml:"issued_office"`
		Nikim          string    `form:"nikim" json:"nikim" xml:"nikim"`
	}

	validate = validator.New()
	validate.RegisterStructValidation(func(structLevel validator.StructLevel) {
		req := structLevel.Current().Interface().(Request)

		if req.Document != "E-KTP" {
			//&& req.Document != "PASSPORT"
			structLevel.ReportError(reflect.ValueOf(req.Document), "Document", "document", "document", "")
		} else {
			/*if (req.Document == "E-KTP" && len(req.NationID) != 16) ||
				(req.Document == "PASSPORT" && len(req.NationID) != 8) {
				structLevel.ReportError(reflect.ValueOf(req.NationID), "NationID", "nation_id", "nation_id", "")
			}*/
			if req.BirthDate == "" {
				structLevel.ReportError(reflect.ValueOf(req.BirthDate), "BirthDate", "birth_date", "birth_date", "")
			} else {
				parse, _ := time.Parse("2006-01-02", req.BirthDate)
				if parse.IsZero() {
					structLevel.ReportError(reflect.ValueOf(req.BirthDate), "BirthDate", "birth_date", "birth_date", "")
				}
			}
			if req.Document == "E-KTP" {
				if len(req.NationID) != 16 {
					structLevel.ReportError(reflect.ValueOf(req.NationID), "NationID", "nation_id", "nation_id", "")
				}
				if req.Fullname == "" {
					structLevel.ReportError(reflect.ValueOf(req.Fullname), "Fullname", "fullname", "fullname", "")
				}
			}
		}
	}, Request{})

	var req Request
	if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{
				"code":   400,
				"status": "Bad Request",
				"error":  err.Error(),
			})
			return
		}
	} else {
		if err := c.ShouldBind(&req); err != nil {
			return
		}
	}

	err := validate.Struct(req)
	if err != nil {
		if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{
				"code":   400,
				"status": "Bad Request",
				"error":  err.Error(),
			})
			return
		}
	}
	tokens := strings.Split(c.Request.Header.Get("Authorization"), " ")
	fmt.Println(tokens[1])
	service := &service.IdentityService{}
	data, err := service.CreateNationIdentity(req.Document, c.GetString("userid"),
		req.NationID, req.Fullname, req.Country, req.BirthDate)
	if err != nil {
		if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusBadRequest, gin.H{
				"code":   400,
				"status": "Bad Request",
				"error":  err.Error(),
			})
			return
		}
	}

	if data["status"].(string) == "SUCCESS" {
		if strings.Contains(c.Request.URL.Path, os.Getenv("URL_API")) {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"status":  "OK",
				"message": data["message"].(string),
				"data": map[string]string{
					"identity_code": data["identity_code"].(string),
				},
			})
			return
		}
	}
}
