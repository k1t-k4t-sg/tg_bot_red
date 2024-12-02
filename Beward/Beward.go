package bw

import (
	//"fmt"
	"errors"
	"log"
	"net/http"
	"io/ioutil"
	"strings"
	"regexp"
	"strconv"
	"time"
)

const (
	SrvCodes	 = "/cgi-bin/srvcodes_cgi?action=get"			//
	SysInfo 	 = "/cgi-bin/systeminfo_cgi?action=get"			//
	IntercomInfo = "/cgi-bin/intercom_cgi?action=get"			//
	DialTest	 = "/cgi-bin/diag_cgi?action=call&Apartment="   // указать квартиру
	StatusDoor	 = "/cgi-bin/intercom_cgi?action=status"		//
	SetParam	 = "/cgi-bin/intercom_cgi?action=set" 			// + &param1=value1&param2=value2
	OpenDoor	 = "/cgi-bin/intercom_cgi?action=maindoor" 		//
	AltDoor		 = "/cgi-bin/intercom_cgi?action=altdoor"		//
	Locked		 = "/cgi-bin/intercom_cgi?action=locked"		//
	LineLevel	 = "/cgi-bin/intercom_cgi?action=linelevel&Apartment=" // указать квартиру
	ApartmentCgi = "/cgi-bin/apartment_cgi?action=get&Number=" // указать квартиру
	Mifare		 = "/cgi-bin/mifare_cgi?action=get"						// параметры mifare
	MifareScan	 = "/cgi-bin/mifare_cgi?action=set&ScanModeActive=on" 	// включить сканирование mifare
	MifareAdd	 = "/cgi-bin/mifare_cgi?action=add"						// добавление метки mifare
	MifareList	 = "/cgi-bin/mifare_cgi?action=list"					// Список ключей

	Rfid		 = "/cgi-bin/rfid_cgi?action=get"						// параметры rfid
	RfidScan	 = "/cgi-bin/rfid_cgi?action=set&RegModeActive=on" 		// включить сканирование rfid

	KKMswitch	 = "/kmnducfg.asp"
	Log	 		 = "/log0.asp"
)

type Intercom struct {
	ID        string
	Vendor    string
	IP    	  string
}

/*
	GetRfid - Параметры сканирования карт mifare

	RegCode=56086
	RegCodeActive=on
	RegKeyValue=00000000000000
	RegModeActive=off

*/
func (i Intercom) GetRfid() (map[string]string, error) {
	rfid_info := HttpIntercomReguest(i.IP+Rfid, "GET")
	if (rfid_info == "404"){
		return nil, errors.New("page Rfid Beward 404")
	}
	rfid_info_arr := strings.Split(rfid_info, "\n")

	var rfid_info_map map[string]string
		rfid_info_map = make(map[string]string)

	for _, val := range rfid_info_arr {
		val_key_value := strings.Split(val, "=")
		if (len(val_key_value) == 2) {
			rfid_info_map[val_key_value[0]] = val_key_value[1]
		}
	}

	return rfid_info_map, nil
}

/*
	GetRfidScan() - Сканирование карт rfid
*/
func (i Intercom) GetRfidScan() string {
	open_status := HttpIntercomReguest(i.IP+RfidScan, "GET")
	if open_status != "OK"{
		return "Error rfid scan"
	}
	return "success"
}


/*
		SetMifareAdd() - Добавление ключа mifare

		Key=806F858A0F6605
		Type=2
		ProtectedMode=on
		CipherIndex=1
		NewCipherEnable=on
		NewCipherIndex=2
		Code=12345678
		Sector=1
		Apartment=65
		Owner=Beward
		AutoPersonalize=on
		Service=on
		Index=2	
*/

/*
	Получение списка ключей
*/
func (i Intercom) GetMifareList() string {
	list := HttpIntercomReguest(i.IP+MifareList, "GET")
	return list
}

func (i Intercom) SetMifareAdd(args ...string) string {
	
	if (len(args) < 1){
		return "Parameters for SET are not specified"
	}
	multiples := ""
	for _, val := range args {
		multiples += "&"+val
	}
	set_mifare := HttpIntercomReguest(i.IP+MifareAdd+multiples, "GET")
	if set_mifare != "OK"{
		return "Error add mifare"
	}
	return "OK"
}
/*
	GetMifareScan() - Запрос на добавление ключей mifare
*/
func (i Intercom) GetMifareScan() string {
	open_status := HttpIntercomReguest(i.IP+MifareScan, "GET")
	if open_status != "OK"{
		return "Error scan mifare"
	}
	return "success"
}
/*
	GetMifare - Параметры сканирования карт mifare

	ScanCode=96507
	ScanCodeActive=on
	ScanKeysIndexes=
	ScanKeysIndexesActive=on
	ScanKeysProtected=off
	AutoCollectKeys=off
	KeyAddCompare=off
	KeyReverse=off
	PersonModeActive=off
	ScanModeActive=off
	AutoExtRfidSync=off
	SmartAutoCollectKeys=off
	SmartAutoCollectAutoDisable=off
	SmartAutoCollectDisableDays=7
*/
func (i Intercom) GetMifare() (map[string]string, error) {
	mifare_info := HttpIntercomReguest(i.IP+Mifare, "GET")
	if (mifare_info == "404"){
		return nil, errors.New("page Mifare Beward 404")
	}
	mifare_info_arr := strings.Split(mifare_info, "\n")

	var mifare_info_map map[string]string
		mifare_info_map = make(map[string]string)

	for _, val := range mifare_info_arr {
		val_key_value := strings.Split(val, "=")
		if (len(val_key_value) == 2) {
			mifare_info_map[val_key_value[0]] = val_key_value[1]
		}
	}

	return mifare_info_map, nil
}
/*
	RequestKKM() - Список коммутации на домофоне
*/

func (i Intercom) RequestKKM() ([]string, error) {

	kkm_sw := HttpIntercomReguest(i.IP+KKMswitch, "GET")

	if !strings.Contains(kkm_sw, "OnSelMode();") {
		return nil, errors.New("KKM not found in the data")
	}

	str_html := strings.Split(kkm_sw, "OnSelMode();")

	deleted_tag := strings.NewReplacer("document.outcfg_frm.du0_"," \nKKM1 E",
									   "document.outcfg_frm.du1_"," \nKKM2 E",
									   "document.outcfg_frm.du2_"," \nKKM3 E",
									   "document.outcfg_frm.du3_"," \nKKM4 E",
									   "document.outcfg_frm.du4_"," \nKKM5 E",
									   "document.outcfg_frm.du5_"," \nKKM6 E",
									   "document.outcfg_frm.du6_"," \nKKM7 E",
									   "document.outcfg_frm.du7_"," \nKKM8 E",
									   "_", "|D",
									   "</script>", "",
									   "</body>", "",
									   "</html>", "",
									   "';", ".Apar",
									   "'", "",
									   "=", " = ",
									   ".value", "").Replace(str_html[2])
	/*
		Разбить полученное значение на масив для поиска строки квартиры
	*/
	return strings.Split(deleted_tag, "\n"), nil
}
/*
	GetLocked() - Запрос состояние дверей 
*/
func (i Intercom) GetLocked() []string {
	locked 		:= HttpIntercomReguest(i.IP+Locked, "GET")
	locked 		= strings.NewReplacer("1", "included", "0", "disconnected").Replace(locked)
	locked_arr 	:= strings.Split(locked, "\n")
	return locked_arr


}
/*
	GetAltDoor() - Запрос на открытие дополнительной двери 
*/
func (i Intercom) GetAltDoor() string {
	open_status := HttpIntercomReguest(i.IP+AltDoor, "GET")
	if open_status != "OK"{
		return "Error executing an opening request"
	}
	return "success Add"
}
/*
	GetOpenDoor() - Запрос на откртыие основной двери
*/
func (i Intercom) GetOpenDoor() string {
	open_status := HttpIntercomReguest(i.IP+OpenDoor, "GET")
	if open_status != "OK"{
		return "Error executing an opening request"
	}
	return "success Door"
}
/*
	SetParamIntercom() - Установка значений 
*/
func (i Intercom) SetParamIntercom(args ...string) string {
	
	if (len(args) < 1){
		return "Parameters for SET are not specified"
	}
	multiples := ""
	for _, val := range args {
		multiples += "&"+val
	}
	set_intercom := HttpIntercomReguest(i.IP+SetParam+multiples, "GET")
	if set_intercom != "OK"{
		return "Error in setting parameters"
	}
	return "success Set mode"
}
/*
	GetStatusDoor() - Статус установки замков и подключенных кнопок
*/
func (i Intercom) GetStatusDoor() map[string]string {
	status_door     := HttpIntercomReguest(i.IP+StatusDoor, "GET")
	status_door 	 = strings.NewReplacer("on", "included", "off", "disconnected").Replace(status_door)
	status_door_arr := strings.Split(status_door, "\n")

	var status_door_map map[string]string
	    status_door_map = make(map[string]string)

	for _, val := range status_door_arr {
		val_key_value := strings.Split(val, "=")
		if (len(val_key_value) == 2) {
			status_door_map[val_key_value[0]] = val_key_value[1]
		}
	}

	return status_door_map
	
}
/*
	GetIntercomInfo() - информация о включенных настройках панели
*/
func (i Intercom) GetIntercomInfo() (map[string]string, error){
	intercom_info     := HttpIntercomReguest(i.IP+IntercomInfo, "GET")
	if intercom_info == ""{
		return nil, errors.New("empty name")
	}
	intercom_info_arr := strings.Split(intercom_info, "\n")

	var intercom_info_map map[string]string
		intercom_info_map = make(map[string]string)

	for _, val := range intercom_info_arr {
		val_key_value := strings.Split(val, "=")
		if (len(val_key_value) == 2) {
			intercom_info_map[val_key_value[0]] = val_key_value[1]
		}
	}

	return intercom_info_map, nil
}
/*
	GetCodes() - Запрос кодов управления панелью
*/
func (i Intercom) GetCodes() map[string]string {
	codes := HttpIntercomReguest(i.IP+SrvCodes, "GET")
	codes_arr := strings.Split(codes, "\n")

	var codes_map map[string]string
		codes_map = make(map[string]string)

	for _, val := range codes_arr {
		val_key_value := strings.Split(val, "=")
		if (len(val_key_value) == 2) {
			codes_map[val_key_value[0]] = val_key_value[1]
		}
	}

	return codes_map
}
/*
	GetSysInfo() - Запрос служебной информации
*/
func (i Intercom) GetSysInfo() map[string]string {
	sys_info  := HttpIntercomReguest(i.IP+SysInfo, "GET")
	sys_info  = strings.NewReplacer("HostName", "Host",
									"DeviceID", "ID",
									"WebVersion", "Web",
									"HardwareVersion", "Hardware",
									"DeviceModel", "Model",
									"DeviceUUID", "UUID",
									"SoftwareVersion", "Software").Replace(sys_info)

	sys_info_arr := strings.Split(sys_info, "\n")

	var sys_info_map map[string]string
		sys_info_map = make(map[string]string)

	for _, val := range sys_info_arr {
		val_key_value := strings.Split(val, "=")
		if (len(val_key_value) == 2) {
			sys_info_map[val_key_value[0]] = val_key_value[1]
		}
	}

	return sys_info_map
}
/*
	GetLineLevel() Запрос уровня трубки в квартире
*/
func (i Intercom) GetLineLevel(apartament string) string {

	// Request failed:Busy

	return HttpIntercomReguest(i.IP+LineLevel+apartament, "GET")
}
/*
	GetDialTest() Тестовый звонок в квартиру
*/
func (i Intercom) GetDialTest(apartament string) string {
	dial := HttpIntercomReguest(i.IP+DialTest+apartament, "GET")

	if dial != "OK"{
		return "Not successful"
	}
	return "Success"
}
/*
	GetApartment() Запрос получения параметров квартиры:
*/
func (i Intercom) GetApartment(apartament string) map[string]string {
	set_intercom := HttpIntercomReguest(i.IP+ApartmentCgi+apartament, "GET")

	set_intercom = strings.NewReplacer("BlockCMS", "Block",
										"PhonesActive", "Active",
										"HandsetUpLevel", "UpLevel",
										"DoorOpenLevel", "OpenLevel").Replace(set_intercom)

	set_intercom_arr := strings.Split(set_intercom, "\n")

	var set_intercom_map map[string]string
		set_intercom_map = make(map[string]string)


	for _, val := range set_intercom_arr {
		val_key_value := strings.Split(val, "=")
		if (len(val_key_value) == 2) {
			set_intercom_map[val_key_value[0]] = val_key_value[1]
		}
	}

	return set_intercom_map

}
/*
	GetLog() Запрос лог журнала 
*/
func (i Intercom) GetLog() (string, string) {
	log := HttpIntercomReguest(i.IP+Log, "GET")

	now  	 := time.Now()
	nsec 	 := now.UnixNano()
	nsec_str := strconv.FormatInt(nsec, 10)

	return nsec_str, log

}
/*
	HttpIntercomReguest() Запрос HTTP к домофону
*/
func HttpIntercomReguest(url, method string) string {
	client := &http.Client{}
	r, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Println(err)
		return ""
	}
	r.Header.Add("Content-Type", "text/plain; charset=utf-8")
	resp, err := client.Do(r)
	if err != nil {
		log.Println(err)
		return ""
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return ""
	}

	if resp.Status == "401 Unauthorized" {
		log.Println(err)
		return "401 Unauthorized"
	}
	
	if resp.Status == "404 Site or Page Not Found" {
		log.Println(err)
		return "404"
	}

	return regexp.MustCompile(`\s*$`).ReplaceAllString(string(body), "")
}

func NewIntercom(ID, Vendor, IP string) (Intercom, error) {

	if (Vendor != "Beward"){
		return Intercom{}, errors.New("Vendor not Found")
	}
    return Intercom{ID, Vendor, IP}, nil
}

/*
	intercom, err := BW.NewIntercom("2175", "Beward", "http://*:*@192.168.0.1:27042/")
	if err != nil {
		log.Fatal(err)
	}

	/**/

	//fmt.Println(intercom.GetCodes())
	//fmt.Println(intercom.GetDialTest("100"))
	//go fmt.Println(intercom.GetSysInfo())
	//go fmt.Println(intercom.GetIntercomInfo())
	//go fmt.Println(intercom.GetLineLevel("100"))
	////fmt.Println(intercom.SetParamIntercom("MainDoorOpenMode=off"))
	//go fmt.Println(intercom.GetStatusDoor()["AltDoorButtincludedPressed"])
	//go fmt.Println(intercom.GetOpenDoor())
	//fmt.Println(intercom.GetAltDoor())
	//fmt.Println(intercom.GetLocked())

