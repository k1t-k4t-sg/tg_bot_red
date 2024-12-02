package main

import (
	"fmt"
	"log"
	"time"
	"io/ioutil"
	"os"
	"sync"
	"regexp"
	"strconv"
	"strings"
	"encoding/json"

	//"./Beward"
	//"./Rosdomofon"
	//rd "rd"
	//bw "bw"

	//"Beward/Beward"
	//"Rosdomofon/Rosdomofon"

	"github.com/kit-kat/rd"
	"github.com/kit-kat/bw"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tidwall/gjson"
)

type InteromUserTG struct {
	Id   string
	KKM  []string
	Chip string
}

var Domofon_List 		string
var Domofon_List_gate 	string
var regexp_level 	= regexp.MustCompile(`(?m)^(L|level|Level|l)\-[0-9]{1,3}$`)
var regexp_correct 	= regexp.MustCompile(`(?m)^(C|correct|Correct|c)\-[0-9]{1,3}$`)
var regexp_intercom = regexp.MustCompile(`(?m)^D{1}\-[0-9]{1,8}$`)
var regexp_dial 	= regexp.MustCompile(`(?m)^Dial\-[0-9]{1,3}$`)
var regexp_addr 	= regexp.MustCompile(`(?m)^[А-Я,а-я]{3,8}$`)
var MifareReq 		= regexp.MustCompile(`(?m)^(Key)([0-9]{1,5})\=.{3,20}$`)
var MifareAddReq 	= regexp.MustCompile(`(?m)^([aA]dd)\-.{3,20}$`)
var User_TG 		= map[int64]InteromUserTG{}

/*
	"6642114797:AAGTzQjMDjDlUaXcwL7hwVWZwvmjnCd_Wdg"
*/
func main() {
	UpdateList()
	bot, err := tgbotapi.NewBotAPI(AuthTokenTG())
	if err != nil {
		log.Panic(err)
	}

	/*
	xType := fmt.Sprintf("%T", bot)
	fmt.Println(xType)
	*/


	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	var list = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("List"),
		),
	)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

		command := update.Message.Text

		if !AuthUserTG(update.Message.From.UserName) {
			msg.Text = "NOT authorized, contact admin"
			log.Printf("Authorized on account %s ", update.Message.From.UserName)
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
			continue
		}

		if command == "/manual" {
			log.Printf("/manual \t%s ", update.Message.From.UserName)
			msg.ParseMode = "markdown"
			var info = ""
			info += "\n - `List` - Список установленных домофонов на сети"
			info += "\n - `D-{}` - Выбрать домофон для дальнейшей операции"
			info += "\n - `Open` - Открыть основную и дополнительную дверь"
			info += "\n - `Full` - Основные переменные домофона"
			info += "\n - `Log`  - log журнал"
			info += "\n - `Key`  - Включить сканирование ключей"
			info += "\n - `Get`  - Запросить последний записанный ключ"
			info += "\n - `L-{}` - Запросить проверку линии до квартиры"
			info += "\n - `Dial-{}` - Выполнить тестовый звонок"
			info += "\n - `{addr}` - Поиск адреса"
			info += "\n - `add-{key}` - Добавить ключ на все панели"
			msg.Text = info

			msg.ReplyMarkup = list
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}

			continue
		}

		if command == "/doc" {
			log.Printf("/doc \t%s ", update.Message.From.UserName)

			msg.Text = "Для получение документации откройте URL \n\n" + "https://github.com/k1t-k4t-sg/bot?tab=readme-ov-file#readme"

			msg.ReplyMarkup = list
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}

			continue

		}

		if command == "/start" {

			log.Printf("/START \t%s ", update.Message.From.UserName)

			msg.ParseMode = "html"
			var info = ""
			info += "Domofon_red\n\n"
			info += "  Просмотр:\n"
			info += "   -уровни квартир\n"
			info += "   -адресации ККМ\n"
			info += "   -IP адрес домофона\n"
			info += "   -системной информации\n\n"
			info += "  Выполнение:\n"
			info += "   -Тестовых звоноков\n"
			info += "   -Открытие двери\n"
			info += "   -Запись ключей\n"
			info += "   -Проверка основных параметров домофонии\n\n"
			info += "Для получение списка установленных домофонов на сети REDCOM наберите <pre>List</pre>\n"
			info += "Для получение справки по командам наберите <pre>/doc</pre>\n"
			info += "Для получение документации по работе с ботом <pre>/help</pre>\n"
			msg.Text = "Success\n\n" + info

			msg.ReplyMarkup = list
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}

			continue

		}
		/*
			Получить последний ключ
		*/
		if command == "get" || command == "Get" {
			msg.ParseMode = "markdown"
			var MifareKey string

			id_chat := update.Message.Chat.ID
			_, ok := User_TG[id_chat]
			if !ok {
				log.Printf("Get => Intercom not selected \t%s ", update.Message.From.UserName)
				msg.Text = "Intercom not selected"

				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
				continue
			}

			id := User_TG[id_chat].Id

			log.Printf("Get \t%v %s ", id, update.Message.From.UserName)

			if !gjson.Get(Domofon_List, id+".URL").Exists() {
				msg.Text = "The intercom with the specified identifier was not found"

				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
				continue
			}

			url := gjson.Get(Domofon_List, id+".URL").String()
			intercom, err := bw.NewIntercom(id, "Beward", url)
			if err != nil {
				log.Fatal(err)
			}

			GetMifareList := intercom.GetMifareList()

			list_map_key := strings.Split(GetMifareList, "\n")
			for _, key := range list_map_key {
				if MifareReq.MatchString(key) {
					key_arr := strings.Split(key, "=")
					MifareKey = key_arr[1]
				}
			}

			if MifareKey == "" {
				fmt.Println("MifareKey not found")
				continue
			}

			msg.Text = "`add-"+MifareKey+"`"

			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}

		}
		/*
			Обновление списка домофонов
		*/
		if command == "update" || command == "Update" {

			log.Printf("UPDATE \t%s ", update.Message.From.UserName)

			UpdateList()

			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			msg.ReplyMarkup = list
			msg.ParseMode = "html"
			msg.Text = "Succes"

			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}

			continue
		}
		/*
			Список установленных домофонов
		*/
		if command == "list" || command == "List" {

			log.Printf("LIST \t%s ", update.Message.From.UserName)

			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			msg.ReplyMarkup = list
			msg.ParseMode = "markdown"

			delete(User_TG, update.Message.Chat.ID)

			var info string
			result := gjson.Get(Domofon_List, "@keys")
			for _, name := range result.Array() {
				id := name.String()
				url := gjson.Get(Domofon_List, id+".URL").String()
				addr := gjson.Get(Domofon_List, id+".Adress").String()
				if url == "Remote IP not found" {
					continue
				}
				//info += "`" + id + "` | " + addr + " \n"
				info += "`D-" + id + "` | " + addr + " \n"
			}
			info += "\nДля поиска введите первые 3 буквы адреса.\n\n"
			info += "Наберите ID домофона для дальнейшей работы\n"
			info += "{D-номер} - Выбрать домофон\n"
			msg.Text = info

			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}

			continue
		}
		/*
			Открытие двери
		*/
		if command == "Open" || command == "open" {

			id_chat := update.Message.Chat.ID
			_, ok := User_TG[id_chat]
			if !ok {
				continue
			}

			id := User_TG[id_chat].Id

			log.Printf("OPEN \t%v %s ", id, update.Message.From.UserName)

			if !gjson.Get(Domofon_List, id+".URL").Exists() {
				continue
			}

			url := gjson.Get(Domofon_List, id+".URL").String()
			intercom, err := bw.NewIntercom(id, "Beward", url)
			if err != nil {
				log.Fatal(err)
			}

			status1 := intercom.GetOpenDoor()
			status2 := intercom.GetAltDoor()

			msg.Text = status1 + "\n" + status2

			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}

			continue
		}
		/*
			Вывод служебной информации
		*/
		if command == "Full" || command == "full" {
			id_chat := update.Message.Chat.ID
			_, ok := User_TG[id_chat]
			if !ok {
				continue
			}

			id := User_TG[id_chat].Id

			log.Printf("FULL \t%v %s ", id, update.Message.From.UserName)

			if !gjson.Get(Domofon_List, id+".URL").Exists() {
				continue
			}

			url := gjson.Get(Domofon_List, id+".URL").String()
			intercom, err := bw.NewIntercom(id, "Beward", url)
			if err != nil {
				log.Fatal(err)
			}

			info_Intercom, err := intercom.GetIntercomInfo()
			if err != nil {
				log.Fatal(err)
			}

			info_Status_Door := intercom.GetStatusDoor()
			info_Codes := intercom.GetCodes()
			info_Sys := intercom.GetSysInfo()

			msg.Text = "\n ***Info Status Door*** \n"

			for key, value := range info_Status_Door {
				msg.Text += key + ": " + value + "\n"
			}

			msg.Text += "\n\n ***Info Intercom*** \n"

			for key, value := range info_Intercom {
				msg.Text += key + ": " + value + "\n"
			}

			msg.Text += "\n\n ***Info Codes*** \n"

			for key, value := range info_Codes {
				msg.Text += key + ": " + value + "\n"
			}

			msg.Text += "\n\n ***Info Sys*** \n"

			for key, value := range info_Sys {
				msg.Text += key + ": " + value + "\n"
			}

			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}

			continue
		}
		/*
			Запрос логов
		*/
		if command == "Log" || command == "log" {
			id_chat := update.Message.Chat.ID
			_, ok := User_TG[id_chat]
			if !ok {
				continue
			}

			id := User_TG[id_chat].Id

			log.Printf("Log \t%v %s ", id, update.Message.From.UserName)

			if !gjson.Get(Domofon_List, id+".URL").Exists() {
				continue
			}

			url := gjson.Get(Domofon_List, id+".URL").String()
			intercom, err := bw.NewIntercom(id, "Beward", url)
			if err != nil {
				log.Fatal(err)
			}

			time_str, info_log := intercom.GetLog()
			patch := "log_time-" + time_str + ".txt"

			/*
				Создание лог журнала
			*/

			file, err := os.Create(patch)
			if err != nil {
				fmt.Println("Unable to create file:", err)
				os.Exit(1)
			}
			file.WriteString(info_log)

			files := tgbotapi.FilePath(patch)
			msgs := tgbotapi.NewDocument(id_chat, files)
			bot.Send(msgs)
			file.Close()

			e := os.Remove(patch)
			if e != nil {
				log.Fatal(e)
			}
			continue
		}
		/*
			Запрос URL
		*/
		if command == "Url" || command == "url" {
			log.Printf("Url \t %s ", update.Message.From.UserName)
			id_chat := update.Message.Chat.ID
			_, ok := User_TG[id_chat]
			if !ok {
				continue
			}
			id_intercom := User_TG[id_chat].Id

			if !gjson.Get(Domofon_List, id_intercom+".URL").Exists() {
				continue
			}

			msg.Text = gjson.Get(Domofon_List, id_intercom+".URL").String()

			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
			continue

		}
		/*
			Добавление ключей
		*/
		if command == "Key" || command == "key" {
			id_chat := update.Message.Chat.ID
			_, ok := User_TG[id_chat]
			if !ok {
				continue
			}

			id := User_TG[id_chat].Id
			chip := User_TG[id_chat].Chip

			log.Printf("Key \t%v %s ", id, update.Message.From.UserName)

			if !gjson.Get(Domofon_List, id+".URL").Exists() {
				continue
			}

			url := gjson.Get(Domofon_List, id+".URL").String()
			intercom, err := bw.NewIntercom(id, "Beward", url)
			if err != nil {
				log.Fatal(err)
			}

			if chip == "MIFARE" {
				status := intercom.GetMifareScan()
				if status == "Error scan mifare" {
					msg.Text = "error"
					continue
				} else {
					msg.Text = "success"
				}
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
				return
			}

			status := intercom.GetRfidScan()
			if status == "Error rfid scan" {
				msg.Text = "error"
				continue
			} else {
				msg.Text = "success"
			}
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
			continue
		}

		/*
			Добавить последний ключ на все домофоны
		*/
		if MifareAddReq.MatchString(command) {

			log.Printf("====> /RECORD KEY \t%s ", update.Message.From.UserName)

			key := strings.Split(command, "-")
			//fmt.Println(key[1])
			var info string
			result := gjson.Get(Domofon_List, "@keys")
			for _, name := range result.Array() {
				//time.Sleep(1 * time.Second)
				go func(name gjson.Result) {

					id := name.String()
					url := gjson.Get(Domofon_List, id+".URL").String()
					if url == "Remote IP not found" {
						return
					}

					intercom, err := bw.NewIntercom(id, "Beward", url)
					if err != nil {
						log.Fatal(err)
						return
					}

					key_add := "Key=" + key[1]
					type_add := "Type=1"
					owner := "Owner=TG_" + update.Message.From.UserName + "(" + string(time.Now().Format("2006_01_02")) + ")"
					//fmt.Println(key_add, type_add, owner)

					// add&Key="+key+"&Type="+switch_type+"&Owner="+owner

					status := intercom.SetMifareAdd(key_add, type_add, owner)

					if status != "OK" {
						fmt.Println(status, url)
						return
					}
					fmt.Println(status, url)
				}(name)
			}

			fmt.Println(info)

		}

		/*
			Выбор панели
		*/
		if regexp_intercom.MatchString(command) {
			IntercomReguest(command, update, msg, bot)
			continue
		}
		/*
			Прозвон квартиры
		*/
		if regexp_level.MatchString(command) {

			apartament := strings.Split(command, "-")[1]

			var Button_Command_apar = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Open"),
					tgbotapi.NewKeyboardButton("Dial-"+apartament),
					tgbotapi.NewKeyboardButton("Level-"+apartament),
					tgbotapi.NewKeyboardButton("Get"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("List"),
					tgbotapi.NewKeyboardButton("Full"),
					tgbotapi.NewKeyboardButton("Log"),
					tgbotapi.NewKeyboardButton("Key"),
				),
			)

			id_chat := update.Message.Chat.ID
			_, ok := User_TG[id_chat]
			if !ok {
				continue
			}
			id_intercom := User_TG[id_chat].Id

			if !gjson.Get(Domofon_List, id_intercom+".URL").Exists() {
				continue
			}

			url := gjson.Get(Domofon_List, id_intercom+".URL").String()

			intercom, err := bw.NewIntercom(id_intercom, "Beward", url)
			if err != nil {
				log.Fatal(err)
			}

			linelevel := intercom.GetLineLevel(apartament)

			linelevel_ADC, err := strconv.ParseFloat(linelevel, 64)
			
			if err == nil {
				linelevel_ADC_volt := (12.0/1024.0)*linelevel_ADC
				linelevel = "("+linelevel+" ADC) ("+strconv.FormatFloat(linelevel_ADC_volt, 'g', 2, 64)+" V)"
			}
			
			get_apartament := intercom.GetApartment(apartament)
			kkm_switch := User_TG[id_chat].KKM

			kkm_apar := " " + apartament + ".Apar"
			var value string
			for _, r := range kkm_switch {
				if strings.Contains(r, kkm_apar) {
					value = r
				}
			}

			log.Printf("LEVEL TEST\t%v %v %s ", id_intercom, apartament, update.Message.From.UserName)

			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			msg.ReplyMarkup = Button_Command_apar
			msg.ParseMode = "html"

			/*

				в положенном 250
				снята 450

				уровень снятия 		= (снята - положена)/2 + положена
				уровень открытия 	= нажата - 50

				Active:on Block:off DoorCode:60732 DoorCodeActive:off Number:1 Phone1:1 Phone2: Phone3: Phone4: Phone5: RegCode:48363 RegCodeActive:on]
			*/
			msg.Text = "<pre>\n"
			msg.Text += "Apartament: " + apartament + "\n"
			msg.Text += "     Level: " + linelevel + "\n"
			msg.Text += "       KKM: " + value + "\n"
			msg.Text += "\n"
			msg.Text += "Active CMS: " + get_apartament["Active"] + "\n"
			msg.Text += " Block CMS: " + get_apartament["Block"] + "\n"
			msg.Text += "    Phone1: " + get_apartament["Phone1"] + "\n"
			msg.Text += "\n"
			msg.Text += "  UP Level: " + get_apartament["UpLevel"] + "\n"
			msg.Text += "Open Level: " + get_apartament["OpenLevel"] + "\n"
			msg.Text += "</pre>\n"

			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}

			continue
		}
		/*
			Тестовый звонок
		*/
		if regexp_dial.MatchString(command) {

			apartament := strings.Split(command, "-")[1]

			log.Printf("DIAL TEST APARTAMENT \t%v %s ", apartament, update.Message.From.UserName)

			id_chat := update.Message.Chat.ID
			_, ok := User_TG[id_chat]
			if !ok {
				continue
			}
			id_intercom := User_TG[id_chat].Id

			if !gjson.Get(Domofon_List, id_intercom+".URL").Exists() {
				continue
			}

			url := gjson.Get(Domofon_List, id_intercom+".URL").String()

			intercom, err := bw.NewIntercom(id_intercom, "Beward", url)
			if err != nil {
				log.Fatal(err)
			}

			IntercomInfo, err := intercom.GetIntercomInfo()
			if err != nil {
				log.Fatal(err)
			}

			msg.Text = "Test call started " + apartament + ", call time = " + IntercomInfo["CallTimeout"] + "\n"
			msg.Text += intercom.GetDialTest(apartament)

			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
		}
		/*
			Поиск адреса
		*/
		if regexp_addr.MatchString(command) {
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			msg.ReplyMarkup = list
			msg.ParseMode = "markdown"

			log.Printf("SEARCH \t%v %s ", command, update.Message.From.UserName)

			delete(User_TG, update.Message.Chat.ID)

			var info string
			result := gjson.Get(Domofon_List, "@keys")

			for _, name := range result.Array() {
				id := name.String()
				url := gjson.Get(Domofon_List, id+".URL").String()
				addr := gjson.Get(Domofon_List, id+".Adress").String()
				if url == "Remote IP not found" {
					continue
				}

				matched, _ := regexp.MatchString(`(?i)`+command, addr)

				if matched {
					info += "`D-" + id + "` | " + addr + " \n"
				}
				
			}

			if len(info) == 0 {
				info += "Не найден"
			}

			info += "\n{D-номер} - Выбрать домофон\n"
			msg.Text = info

			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}

			continue
		}

	}
}

func IntercomReguest(command string, update tgbotapi.Update, msg tgbotapi.MessageConfig, bot *tgbotapi.BotAPI) {

	
	var Button_Command = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Open"),
			tgbotapi.NewKeyboardButton("List"),
			tgbotapi.NewKeyboardButton("Full"),
			tgbotapi.NewKeyboardButton("Log"),
			tgbotapi.NewKeyboardButton("Key"),
			tgbotapi.NewKeyboardButton("Get"),
		),
	)

	/**/
	fmt.Println(Button_Command)

	id := strings.Split(command, "-")[1]
	id_chat := update.Message.Chat.ID

	log.Printf("===> SELECTED \t%v %s ", id, update.Message.From.UserName)

	if !gjson.Get(Domofon_List, id+".URL").Exists() {
		return
	}

	vendor 		:= gjson.Get(Domofon_List, id+".Vendor").String()
	url 		:= gjson.Get(Domofon_List, id+".URL").String()
	addr 		:= gjson.Get(Domofon_List, id+".Adress").String()
	FlatStart 	:= gjson.Get(Domofon_List, id+".FlatStart").String()
	FlatEnd 	:= gjson.Get(Domofon_List, id+".FlatEnd").String()
	

	if vendor == "QTECH; QDB-27C-H" {
		msg.Text = "QTECH; QDB-27C-H не поддерживается"

		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
		return
	}
	

	intercom, err := bw.NewIntercom(id, "Beward", url)
	if err != nil {
		fmt.Println(err)
	}

	log.Println(intercom) 

	/*
	ch := make(chan int)
   go func() {
       result := someLongComputation()
       ch <- result
   }()
   */

   chan_info_intercom 	:= make(chan map[string]string)
   chan_kkm_switch    	:= make(chan []string)
   chan_status_door   	:= make(chan []string)
   chan_sys_info 	   	:= make(chan map[string]string)
   chan_mifare		   	:= make(chan map[string]string)
   chan_GetMifareList 	:= make(chan string)
   chan_error_log	   	:= make(chan string)
   chan_rfid		   	:= make(chan map[string]string)


   var wg sync.WaitGroup
   //ctx, cancel := context.WithCancel(context.Background()) 
	//defer cancel()
   //timer := time.After(5 * time.Second) 

   wg.Add(1)
   go func() {
	   defer wg.Done()
	   result, err := intercom.GetIntercomInfo()
	   if err != nil {
		   log.Println(err)
		   msg.Text = "Not available"
		   if _, err := bot.Send(msg); err != nil {
			   log.Panic(err)
		   }
		   chan_info_intercom <- nil
		   return
	   }
	   chan_info_intercom <- result
	   return
   }()

   wg.Add(1)
   go func() {
	   defer wg.Done()
	   result, err := intercom.RequestKKM()
	   if err != nil {
		   msg.Text = "KKM not found in the data"
		   if _, err := bot.Send(msg); err != nil {
			   log.Panic(err)
		   }
		   chan_kkm_switch <- nil
		   return
	   }
	   chan_kkm_switch <- result
	   return
   }()

   wg.Add(1)
   go func() {
	   defer wg.Done()
	   chan_status_door <- intercom.GetLocked()
	   return
   }()

   wg.Add(1)
   go func(){
	   defer wg.Done()
	   chan_sys_info <- intercom.GetSysInfo()
	   return
   }()

   wg.Add(1)
   go func(){
	   defer wg.Done()
	   result, err := intercom.GetMifare()
	   if err != nil && result == nil {
		   chan_mifare <- nil
		   return
	   }
	   chan_mifare <- result
	   return
   }()

   wg.Add(1)
   go func(){
	   defer wg.Done()
	   
	   result, err := intercom.GetRfid()
	   if err != nil {
		   chan_rfid <- nil
		   return
	   }
	   chan_rfid <- result
	   return
   }()

   wg.Add(1)
   go func(){
	   defer wg.Done()
	   chan_GetMifareList <- intercom.GetMifareList()
	   return
   }()

   wg.Add(1)
   go func(){
	   defer wg.Done()
	   _, info_log := intercom.GetLog()
	   chan_error_log <- strconv.Itoa(strings.Count(info_log, "ERROR"))
	   return
   }()


	info_intercom	:= <-chan_info_intercom
	kkm_switch	 	:= <-chan_kkm_switch
	status_door	 	:= <-chan_status_door
	sys_info	 	:= <-chan_sys_info
	mifare	 		:= <-chan_mifare
	GetMifareList	:= <-chan_GetMifareList
	error_log	 	:= <-chan_error_log
	rfid	 		:= <-chan_rfid

	if info_intercom == nil {
		msg.Text = "The operation failed"
		if _, err := bot.Send(msg); err != nil {
		   log.Panic(err)
		}
	}

   	wg.Wait()

	User_TG[id_chat] = InteromUserTG{Id: id, KKM: kkm_switch, Chip: ""}

	DoorOpenLevel 	:= info_intercom["DoorOpenLevel"]
	HandsetUpLevel 	:= info_intercom["HandsetUpLevel"]
	DoorCode 		:= info_intercom["DoorCode"]
	DoorCodeActive 	:= info_intercom["DoorCodeActive"]
	AltDoorOpened 	:= "not installed"

	if len(status_door) == 2 {
		AltDoorOpened = status_door[1]
	}

	MainDoorOpened  := status_door[0]
	Uptime 			:= sys_info["UpTime"]

	var ScanCode string
	var ScanCodeActive string
	var KeyReverse string
	var count string

	if rfid != nil {

		ScanCode 			= rfid["RegCode"]
		ScanCodeActive 		= rfid["RegCodeActive"]
		KeyReverse			= "no RFID support"
		count 				= "no RFID support"
		User_TG[id_chat] 	= InteromUserTG{Id: id, KKM: kkm_switch, Chip: "RFID"}
	}

	if mifare != nil {
		ScanCode 		= mifare["ScanCode"]
		ScanCodeActive 	= mifare["ScanCodeActive"]
		KeyReverse 		= mifare["KeyReverse"]
		count 			= strconv.Itoa(strings.Count(GetMifareList, "Key"))
		User_TG[id_chat] = InteromUserTG{Id: id, KKM: kkm_switch, Chip: "MIFARE"}
	}

	var info string
	info += "<pre>\n"
	info += "ID:     " + id + "\n"
	info += "Addr:   " + addr + "\n"
	info += "Uptime: " + Uptime + "\n"
	info += "Error:  " + error_log + "\n"
	info += "Label:  " + count + " key" + "\n"
	info += "\n"
	info += "   Numbering: " + FlatStart + "-" + FlatEnd + "\n"
	info += " AnswerLevel: " + HandsetUpLevel + "\n"
	info += "   DoorLevel: " + DoorOpenLevel + "\n"
	info += "\n"
	info += "    ScanCode: " + ScanCode + "=>" + ScanCodeActive + "\n"
	info += "  KeyReverse: " + KeyReverse + "\n"
	info += "\n"
	info += "    DoorCode: " + DoorCode + "=>" + DoorCodeActive + "\n"
	info += "    MainDoor: " + MainDoorOpened + "\n"
	info += "     AltDoor: " + AltDoorOpened + "\n"
	info += "\nВведите:\n"
	info += "{L-номер} - Запрос уровня квартиры\n"
	info += "{Open}    - Открыть дверь\n"
	info += "{Full}    - Установленные параметры\n"
	info += "{Log}     - Журнал сообщений домофона\n"
	info += "</pre>\n"

	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	msg.ReplyMarkup = Button_Command
	msg.ParseMode = "html"
	msg.Text = info

	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}

}

func UpdateList() {

	id, user, pass := AuthTokenRosdomofon()

	rosdomofon := rd.TokenGET(id, user, pass)

	//map[string]rd.Domofon
	rosdomofon_obj := rosdomofon.Connection()

	file_gate, err := os.Open("gate.json")
	if err != nil {
		fmt.Println("Unable to create file:", err)
		os.Exit(1)
	}
	defer file_gate.Close()
	ByteAuth, _ := ioutil.ReadAll(file_gate)
	Domofon_List_gate = string(ByteAuth)

	//map[string]rd.Domofon
	var gate_obj map[string]rd.Domofon
	err = json.Unmarshal([]byte(Domofon_List_gate), &gate_obj)
	if err != nil {
		panic(err)
	}

	/*
		объедененный масcив gate_obj и rosdomofon_obj
	*/
	united := make(map[string]rd.Domofon)

	for key, value := range gate_obj {
		united[key] = value
	}

	for key1, value1 := range rosdomofon_obj {
		united[key1] = value1
	}

	jsonStr, err := json.Marshal(united)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
	file, err := os.Create("List.json")
	if err != nil {
		fmt.Println("Unable to create file:", err)
		os.Exit(1)
	}
	defer file.Close()
	file.WriteString(string(jsonStr))
	Domofon_List = string(jsonStr)

	/*
		file, err = os.Open("List.json")
		if err != nil {
			fmt.Println("Unable to create file:", err)
			os.Exit(1)
		}
		defer file.Close()
		ByteAuth, _ = ioutil.ReadAll(file)
		Domofon_List = string(ByteAuth)


	*/
}

func AuthUserTG(user string) bool {

	auth, err := os.Open("auth.json")
	if err != nil {
		fmt.Println(err)
	}
	defer auth.Close()
	ByteAuth, _ := ioutil.ReadAll(auth)
	if !gjson.ValidBytes(ByteAuth) {
		fmt.Println("ERROR ==> auth.json")
		return false
	}

	if gjson.GetBytes(ByteAuth, "user.#(=="+user+"))").Exists() {
		return true
	} else {
		return false
	}
}

func AuthTokenTG() string {

	auth, err := os.Open("auth.json")
	if err != nil {
		fmt.Println(err)
	}
	defer auth.Close()
	ByteAuth, _ := ioutil.ReadAll(auth)

	token := gjson.Get(string(ByteAuth), "admin").String()
	return token
}

func AuthTokenRosdomofon() (string, string, string) {

	auth, err := os.Open("auth.json")
	if err != nil {
		fmt.Println(err)
	}
	defer auth.Close()
	ByteAuth, _ := ioutil.ReadAll(auth)

	id   := gjson.Get(string(ByteAuth), "auth.0.client_id").String()
	user := gjson.Get(string(ByteAuth), "auth.0.username").String()
	pass := gjson.Get(string(ByteAuth), "auth.0.password").String()	
	
	return id, user, pass
}
