package rd

import (
        //"fmt"
        "net/http"
        "net/url"
        "io/ioutil"
        "os"
        "strings"
        "log"
        "regexp"

        "github.com/tidwall/gjson"
)

const (
        Regexp_URL = `(http:\/\/[a-zA-Z]*\:[a-zA-Z0-9]*@)([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})\:([0-9]{1,5})?`
)

const (
        ApiUrl    = "https://rdba.rosdomofon.com"
        Entrances = "/rdas-service/api/v1/entrances/"
        Resource  = "/authserver-service/oauth/token"
)

type RDA struct {
        Token string
}

type Domofon struct {
        ID        string `json:"ID"`
        Vendor    string `json:"Vendor"`
        Adress    string `json:"Adress"`
        URL       string `json:"URL"`
        FlatStart int64  `json:"FlatStart"`
        FlatEnd   int64  `json:"FlatEnd"`
}

type Abonent struct {
        SOFT_ID               string
        HARD_ID               string
        Intercom              string
        SoftwareIntercomOwner string
        HardwareIntercomOwner string
}

/*
        Connection() - Список подключенных домофонов с статусовм online
*/
func (r RDA) Connection() map[string]Domofon {
	json_connection := HttpRDA_Reguest(r.Token, "GET", ApiUrl+Entrances+"?&size=1000")

        result := gjson.Get(json_connection, "content")

        //fmt.Println(result)

        OBJEST_Domofon := map[string]Domofon{}

        for _, name := range result.Array() {
                id               := gjson.Get(name.String(), "id").String()
                street   := gjson.Get(name.String(), `address.street.name`).String()
                building := gjson.Get(name.String(), `address.house.building`).String()
                house    := gjson.Get(name.String(), `address.house.number`).String()
                housing    := gjson.Get(name.String(), `address.house.housing`).String()
                entrance := gjson.Get(name.String(), `address.entrance.number`).String()
                

                flatStart := gjson.Get(name.String(), `address.entrance.flatStart`).Int()
                flatEnd   := gjson.Get(name.String(), `address.entrance.flatEnd`).Int()

                if building != ""{
                        building = " стр."+building
                }
                if housing != ""{
                        housing = " к."+housing
                }

                OBJEST_Domofon[id] = Domofon{Vendor: gjson.Get(name.String(), `rda.intercomType.name`).String(),
                        Adress:    street + " д." + house + building + housing + " Под " + entrance,
                        URL:       SearchIP_URL(gjson.Get(name.String(), `rda.configStr`).String()),
                        FlatStart: flatStart,
                        FlatEnd:   flatEnd}
        }

        return OBJEST_Domofon


}

/*
        Flats() - Список подключенных квартир в реестре росдомофона

        ID string - ID домофона
*/
func (r RDA) Flats(ID string) map[string]Abonent {

        json_connection := HttpRDA_Reguest(r.Token, "GET", ApiUrl+Entrances+ID+"/flats")
        result                  := gjson.Get(json_connection, "@reverse")

        Flat := map[string]Abonent{}

        for _, name := range result.Array() {
                isVirtual       := gjson.Get(name.String(), "isVirtual").String()
                flat            := gjson.Get(name.String(), "address.flat").String()
				idSOFT          := gjson.Get(name.String(), "softwareIntercomOwner.id").String()
                idHARD          := gjson.Get(name.String(), "hardwareIntercomOwner.id").String()
                softwareIntercomOwner := gjson.Get(name.String(), "softwareIntercomOwner.phone").String()
                hardwareIntercomOwner := gjson.Get(name.String(), "hardwareIntercomOwner.phone").String()

                Flat[flat] = Abonent{SOFT_ID:                           idSOFT,
                                                         HARD_ID:               idHARD,
                                                         Intercom:              isVirtual,
                                                         SoftwareIntercomOwner: softwareIntercomOwner,
                                                         HardwareIntercomOwner: hardwareIntercomOwner}
        }

        return Flat
}

func TokenGET(client_id, username, password string) RDA {

        data := url.Values{}

        data.Set("grant_type", "password")
        data.Set("client_id",  client_id)
        data.Set("username",   username)
        data.Set("password",   password)

        u, _    := url.ParseRequestURI(ApiUrl)
        u.Path   = Resource
        urlStr  := u.String()

        client := &http.Client{}
        r, _ := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(data.Encode()))
        r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

        resp, _ := client.Do(r)
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
                log.Fatalln(err)
        }

        value := gjson.Get(string(body), "access_token")
        if value.String() != "" {
                file, err := os.Create("token.txt")
                if err != nil {
                        //fmt.Println("Unable to create file:", err)
                        os.Exit(1)
                }
				defer file.Close()
                file.WriteString(value.String())
                return RDA{value.String()}
        }

        return RDA{}
}

func HttpRDA_Reguest(Token, method, url string) string {
        client := &http.Client{}
        r, _ := http.NewRequest(method, url, nil)
        r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
        r.Header.Add("Authorization", "Bearer "+Token)

        resp, _ := client.Do(r)
        body, err := ioutil.ReadAll(resp.Body)

        if err != nil {
                log.Fatalln(err)
        }


        if resp.Status == "401 Unauthorized" {
                return "Unauthorized"
        }

        return string(body)
}

func SearchIP_URL(URL string) string {
        //fmt.Println(URL)
        IP_Regexp := regexp.MustCompile(Regexp_URL)
        IP_Port := IP_Regexp.FindString(URL)
        if IP_Port != "" {
                return IP_Port
        }
        return "Remote IP not found"
}