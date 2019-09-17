package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/vds/go-resman/pkg/database"
	"github.com/vds/go-resman/pkg/middleware"
	"github.com/vds/go-resman/pkg/models"
	"net/http"
)

type RestaurantController struct{
	database.Database
}

func NewRestaurantController(db database.Database) *RestaurantController{
	resController:=new(RestaurantController)
	resController.Database=db
	return resController
}

func(r *RestaurantController)GetNearBy(c *gin.Context){
	var location models.Location
	err:=c.ShouldBindJSON(&location)
	if err!=nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	var jsonData=[]struct{
		Name string `json:"name" binding:"required"`
	}{}
	stringData,err:=r.ShowNearBy(&location)
	if stringData==""{
		c.JSON(http.StatusOK, gin.H{
			"msg":"No restaurants to show",
		})
		return
	}
	_=json.Unmarshal([]byte(stringData),&jsonData)

	c.JSON(http.StatusOK,jsonData)
}

func(r *RestaurantController)GetRestaurants(c *gin.Context){
	value,_:=c.Get("userAuth")
	userAuth:=value.(*models.UserAuth)
	jsonData:=&[]models.RestaurantOutput{}
	var stringData string
	var err error
	stringData, err = r.ShowRestaurants(userAuth)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if stringData==""{
		c.JSON(http.StatusOK, gin.H{
			"error":"No restaurants to show",
		})
		return
	}
	_=json.Unmarshal([]byte(stringData),jsonData)

	c.JSON(http.StatusOK,jsonData)
}

func (r *RestaurantController)AddRestaurant(c *gin.Context){
	value,_:=c.Get("userAuth")
	userAuth:=value.(*models.UserAuth)
	var restaurant models.Restaurant
	err:=c.ShouldBindJSON(&restaurant)
	if err!=nil {
		fmt.Print(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	restaurant.CreatorID=userAuth.ID
	err=r.InsertRestaurant(&restaurant)
	if err!=nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"msg":"Restaurant added",
	})
}

func (r *RestaurantController)EditRestaurant(c *gin.Context){
	value,_:=c.Get("userAuth")
	userAuth:=value.(*models.UserAuth)
	if userAuth.Role!=middleware.Admin && userAuth.Role!=middleware.SuperAdmin{
		c.Status(http.StatusUnauthorized)
		return
	}
	res,_:=c.Get("restaurantID")
	resID:=res.(int)
	var restaurant models.RestaurantOutput
	restaurant.ID=resID
	err:=c.ShouldBindJSON(&restaurant)
	if err!=nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	err=r.UpdateRestaurant(&restaurant)
	if err!=nil{
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg": "Restaurant Updated Successfully",
	})
}

func (r *RestaurantController)DeleteRestaurants(c *gin.Context){
	value,_:=c.Get("userAuth")
	userAuth:=value.(*models.UserAuth)
	var resID struct {
		IDArr []int	`json:"idArr" binding:"required"`
	}
	err:=c.ShouldBindJSON(&resID)
	if err!=nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	err=r.RemoveRestaurants(userAuth,resID.IDArr...)
	if err!=nil{
		fmt.Printf("err is %v",err)
		if err!=database.ErrInternal{
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"msg": "Restaurants deleted Successfully",
	})
}
func(r *RestaurantController)GetOwnerRestaurants(c * gin.Context){
	ownerID := c.Param("ownerID")
	value,_:=c.Get("userAuth")
	userAuth:=value.(*models.UserAuth)
	var err error
	if userAuth.Role==middleware.Admin {
		err = r.CheckOwnerCreator(userAuth.ID,ownerID)
		if err != nil {
			if err!=database.ErrInternal {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.Status(http.StatusInternalServerError)
			return
		}
	}
	jsonData:=&[]models.RestaurantOutput{}
	ownerAuth:=models.UserAuth{
		ID:   ownerID,
		Role: middleware.Owner,
	}
	var stringData string
	stringData, err = r.ShowRestaurants(&ownerAuth)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if stringData==""{
		c.JSON(http.StatusOK, gin.H{
			"error":"No restaurants to show",
		})
		return
	}
	_=json.Unmarshal([]byte(stringData),jsonData)

	c.JSON(http.StatusOK,jsonData)
}

func (r *RestaurantController)GetAvailableRestaurants(c *gin.Context){
	value,_:=c.Get("userAuth")
	userAuth:=value.(*models.UserAuth)
	jsonData:=&[]models.RestaurantOutput{}
	stringData,err:=r.ShowAvailableRestaurants(userAuth)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if stringData==""{
		c.JSON(http.StatusOK, gin.H{
			"msg":"No restaurants to show",
		})
		return
	}
	_=json.Unmarshal([]byte(stringData),jsonData)

	c.JSON(http.StatusOK,jsonData)
}

func (r *RestaurantController)AddOwnerForRestaurants(c *gin.Context){
	ownerID := c.Param("ownerID")
	value,_:=c.Get("userAuth")
	userAuth:=value.(*models.UserAuth)
	var err error
	if userAuth.Role==middleware.Admin {
		err = r.CheckOwnerCreator(userAuth.ID,ownerID)
		if err != nil {
			if err!=database.ErrInternal {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.Status(http.StatusInternalServerError)
			return
		}
	}
	var resID struct {
		IDArr []int	`json:"idArr" binding:"required"`
	}
	err=c.ShouldBindJSON(&resID)
	if err!=nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	err=r.InsertOwnerForRestaurants(userAuth,ownerID,resID.IDArr...)
	if err!=nil{
		if err!=database.ErrInternal{
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"msg": "Owner added Successfully",
	})
}