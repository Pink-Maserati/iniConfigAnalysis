package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
)

//测试提交
//ini配置文件解析器

//MysqlConfig Mysql配置结构体
type MysqlConfig struct {
	Address  string `ini:"address"`
	Port     int    `ini:"port"`
	Username string `ini:"username"`
	Password string `ini:"password"`
}

//RedisConfig Redis配置结构体
type RedisConfig struct {
	Host     string `ini:"host"`
	Port     int    `ini:"port"`
	Password string `ini:"password"`
	Database int    `ini:"database"`
	Test     bool   `ini:"test"`
}

// Config 结构体
type Config struct {
	MysqlConfig `ini:"mysql"`
	RedisConfig `ini:"redis"`
}

func loadIni(fileName string, data interface{}) (err error) {
	//1.参数校验：
	//1.1传进来的data参数必须是指针类型(因为需要在函数中对其进行赋值，值类型/引用类型的区别)
	t := reflect.TypeOf(data)
	// fmt.Printf("测试:%v %v\n", t, t.Kind())
	if t.Kind() != reflect.Ptr {
		err = errors.New("data should be a pointer") //新创建一个错误
		return
	}
	//1.2 传进来的data参数必须是结构体类型指针（配置文件文件不止一个字段，需要将配置文件中的各种键值对赋值给结构体字段）
	if t.Elem().Kind() != reflect.Struct {
		err = errors.New("data should be a struct pointer")
		return
	}

	//2.读文件得到字节类型数据
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return
	}
	//string(b)将字节数组转换为字符串
	lineSlice := strings.Split(string(b), "\n")
	// fmt.Printf("%#v\n", lineSlice)
	//3.一行一行读数据
	var structName string
	for idx, line := range lineSlice {
		//去掉字符串首尾空格
		line = strings.TrimSpace(line)
		//如果是空行，直接过滤
		if len(line) == 0 {
			continue
		}
		//如果是注释，跳过
		if strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		//如果是[]表示是节 section
		if strings.HasPrefix(line, "[") {
			if line[0] != '[' || line[len(line)-1] != ']' {
				err = fmt.Errorf("line:%d syntax error", idx+1)
				return

			}
			//[] [...] 若...为空也有问题，拿到中间的内容
			sectionName := strings.TrimSpace(line[1 : len(line)-1])
			if len(sectionName) == 0 {
				err = fmt.Errorf("line:%d syntax error", idx+1)
				return

			}
			//根据字符串sectionName的内容去data里面根据反射找对应的结构体
			for i := 0; i < t.Elem().NumField(); i++ {
				field := t.Elem().Field(i)
				if sectionName == field.Tag.Get("ini") {
					structName = field.Name
					fmt.Printf("找到%s对应的嵌套结构体%s\n", sectionName, structName)
				}

			}

		} else {
			//如果不是[]开头的就是内容， host=127.0.0.1 用=分割键值对
			if !strings.Contains(line, "=") || strings.HasPrefix(line, "=") {
				err = fmt.Errorf("line:%d syntax error", idx+1)
				return
			}

			//以=分割这一行，等号左边是key,等号右边是value
			index := strings.Index(line, "=")
			key := strings.TrimSpace(line[:index])
			value := strings.TrimSpace(line[index+1:])
			//根据structName去把data里面对应的嵌套结构体给取出来
			v := reflect.ValueOf(data)
			sValue := v.Elem().FieldByName(structName) //拿到嵌套结构体的值信息
			sType := sValue.Type()                     //拿到嵌套结构体的类型信息
			if sType.Kind() != reflect.Struct {
				err = fmt.Errorf("data中的%s字段应该是一个结构体", structName)
				return
			}
			var fieldName string
			var fileType reflect.StructField

			//遍历嵌套结构体的每一个字段 判断tag是不是等于key
			for i := 0; i < sValue.NumField(); i++ {
				field := sType.Field(i) //tag信息是存储在类型信息中
				fileType = field
				if field.Tag.Get("ini") == key {
					//找到对应的字段
					fieldName = field.Name
					break
				}
			}
			//4.如果key=tag，给这个字段赋值
			//根据fieldName去取这个字段
			if len(fieldName) == 0 {
				//结构体中找不到对应的字段
				continue
			}
			fileObj := sValue.FieldByName(fieldName)
			//对其赋值
			fmt.Println(fieldName, fileType.Type.Kind())
			switch fileType.Type.Kind() {
			case reflect.String:
				fileObj.SetString(value)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				var valueInt int64
				valueInt, err = strconv.ParseInt(value, 10, 64)
				if err != nil {
					err = fmt.Errorf("line:%d value type error", idx+1)
					return

				}
				fileObj.SetInt(valueInt)

			case reflect.Bool:
				var valueBool bool
				valueBool, err = strconv.ParseBool(value)
				if err != nil {
					err = fmt.Errorf("line:%d value type error", idx+1)
					return

				}
				fileObj.SetBool(valueBool)
			case reflect.Float32, reflect.Float64:
				var valueFloat float64
				valueFloat, err = strconv.ParseFloat(value, 64)
				if err != nil {
					err = fmt.Errorf("line:%d value type error", idx+1)
					return

				}
				fileObj.SetFloat(valueFloat)

			}

		}

	}
	return nil

}
func main() {

	var cfg Config
	// var x = new(int)
	err := loadIni("./conf.ini", &cfg)
	if err != nil {

		fmt.Printf("load ini failed,err:%v\n", err)

		return
	}
	fmt.Printf("%#v\n", cfg)
	fmt.Println(cfg.MysqlConfig)

}
