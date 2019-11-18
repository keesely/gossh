// assh.go kee > 2019/11/08

package assh

import (
	"fmt"
	"github.com/keesely/kiris"
	"github.com/keesely/kiris/hash"
	"log"
)

type Assh struct {
	data   *kiris.Yaml
	dbFile string
}

var passwd string

func NewAssh() *Assh {
	cFile := cPath + "/servers.ydb"
	passwd = GetPasswd()

	var (
		_data = []byte{}
		pwd   = ""
	)
	if kiris.FileExists(cFile) {
		// 导入数据文件
		_data, pwd = getDataContents(cFile)
		if pwd != hash.Md5(passwd) {
			log.Fatal("Access denied for password. run assh account [your password]")
		}
	}
	data := kiris.NewYaml(_data)

	return &Assh{data, cFile}
}

// 获取数据文件内容
func getDataContents(datafile string) ([]byte, string) {
	content, err := kiris.FileGetContents(datafile)
	if err != nil {
		log.Fatal(err)
	}
	// 解码还原数据
	_c := kiris.AESDecrypt(content[0:len(content)-32], passwd, "cbc")
	if _c == "" && string(content) != "" {
		log.Fatal("the password is error")
	}
	return []byte(_c), string(content[len(content)-32:])
}

func (c *Assh) ListServers() {
	fmt.Println(kiris.StrPad("", "=", 100, 0))
	fmt.Printf("| %-20s | %-20s | %-50s |\n", "Group Name", "Server Name", "Server Host")
	fmt.Println(kiris.StrPad("", "-", 100, 0))
	data := c.data.Get("")
	for g, ss := range data.(map[string]interface{}) {
		for n, _s := range ss.(map[string]interface{}) {
			s := &Server{}
			kiris.ConverStruct(_s.(map[string]interface{}), s, "yaml")
			passwd := kiris.Ternary(s.Password != "", "yes", "no").(string)
			fmt.Printf("> %-20s | %-20s | %s@%s:%d (password:%s) \n", g, n, s.User, s.Host, s.Port, passwd)
		}
	}
	//c.save()
	fmt.Println(kiris.StrPad("", "=", 100, 0))
}

func (c *Assh) AddServer(name string, server *Server) {
	name = "default." + name
	c.data.Set(name, &server)
	// 保存
	c.save()
	fmt.Printf("Server [%s] add success!\n", name)
	//fmt.Println("Save to ", saveFs)
}

func (c *Assh) GetServer(name string) *Server {
	if server := c.data.Get(name); server != nil {
		ss := &Server{}
		kiris.ConverStruct(server.(map[string]interface{}), ss, "yaml")
		return ss
		/*
			return &Server{
				Name:     s.Name,
				Host:     s.Host,
				Port:     s.Port,
				User:     s.User,
				Options:  s.Options,
				Password: s.Password,
				PemKey:   s.PemKey,
			}
		*/
	}
	return nil
}

func (c *Assh) DelServer(name string) {
	c.data.Set(name, nil)
	c.save()
	fmt.Println("删除成功: ", name)
	return
}

func (c *Assh) save() {
	str, _ := c.data.SaveToString()
	// 加密
	save := kiris.AESEncrypt(string(str), passwd, "cbc")
	//fmt.Println("encrypt: ", string(kiris.Base64Decode(save)))
	content := string(save) + hash.Md5(passwd)

	saveFs := kiris.RealPath(c.dbFile)
	//if e := c.data.SaveAs(saveFs); e != nil {
	if e := kiris.FilePutContents(saveFs, content, 0); e != nil {
		log.Fatal(e)
	}
}