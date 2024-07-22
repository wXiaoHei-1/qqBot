package server

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

const dataTxtPath string = "The path to your corpus file"

var (
	idiomMap     map[string][]string
	currentIdiom string
)

// NewIdiomMap 初始化词库
func NewIdiomMap() {
	idiomMap = make(map[string][]string)
	// 创建一个map来存储成语
	chengYuMap := make(map[string][]string)

	// 打开文件
	file, err := os.Open(dataTxtPath)
	if err != nil {
		fmt.Println("打开文件失败:", err)
		panic(err)
	}
	defer file.Close()

	// 创建一个缓冲区读取器
	reader := bufio.NewReader(file)

	// 逐行读取文件
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		// 去掉换行符
		line = strings.TrimSpace(line)
		// 将成语按逗号分割
		chengYus := strings.Split(line, ",")
		// 遍历每个成语,并将其加入到map中
		for _, chengYu := range chengYus {
			key := GetFirstChineseChar(chengYu)
			chengYuMap[key] = append(chengYuMap[key], chengYu)
		}
	}
	idiomMap = chengYuMap
}

// ChengYvInterlocking 成语接龙游戏进行逻辑
func ChengYvInterlocking(idiom string) (string, bool) {
	//去除空格
	idiom = strings.TrimSpace(idiom)
	if idiom == "" {
		return "输入不能为空,请重新输入。", false
	}
	//判断是否全为中文
	isChineseChar := isAllChineseCharacters(idiom)
	if !isChineseChar {
		return "您这边输入的好像不是成语，请重试", false
	}
	//判断是否为四字成语
	if len(idiom) < 12 {
		return "您输入的不是四字词语请重新输入", false
	}
	//判断是否是一句新的开局游戏，如果不是检查用户输入是否正确
	if currentIdiom != "" {
		flag := checkIdiom(currentIdiom, idiom)
		if !flag {
			return "您输入的成语不符合游戏规则,请重新输入。", false
		}
	}
	//查询符合游戏规则的下一个单词
	nextIdiom := FindNextIdiom(idiom)
	//没有找到，则将记录清空并返回游戏技术标志true
	if nextIdiom == "" {
		currentIdiom = ""
		return fmt.Sprintf("没有找到可以接上'%s'的成语，恭喜你获得游戏胜利。\n", idiom), true
	}
	currentIdiom = nextIdiom
	return nextIdiom, false
}

// ResetCurrentIdiom 清空机器上次回答记录
func ResetCurrentIdiom() {
	currentIdiom = ""
}

// checkIdiom 判断用户回答是否正确
func checkIdiom(idiom1 string, idiom2 string) bool {
	idiom2FirstChar := GetFirstChineseChar(idiom2)
	idiom1LastChar := GetLastChineseChar(idiom1)
	if strings.EqualFold(idiom2FirstChar, idiom1LastChar) {
		return true
	}
	return false
}

// FindNextIdiom 查询词库符合条件的单词
func FindNextIdiom(idiom string) string {
	firstChineseChar := GetLastChineseChar(idiom)
	if value, ok := idiomMap[firstChineseChar]; ok {
		randomNum := rand.Intn(len(value))
		return value[randomNum]
	}
	return ""
}

// GetFirstChineseChar 获取中文字符串的第一个字符
func GetFirstChineseChar(s string) string {
	r, _ := utf8.DecodeRuneInString(s)
	return string(r)
}

// GetLastChineseChar 获取中文字符串的最后一个字符
func GetLastChineseChar(s string) string {
	_, size := utf8.DecodeLastRuneInString(s)
	return s[len(s)-size:]
}

// isAllChineseCharacters 判断字符是否全为中文字符
func isAllChineseCharacters(s string) bool {
	for _, r := range s {
		if !unicode.Is(unicode.Han, r) {
			return false
		}
	}
	return true
}
