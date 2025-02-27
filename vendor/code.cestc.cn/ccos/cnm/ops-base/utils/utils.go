package utils

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// GetIntranetIp 获取本机内网IP
func GetIntranetIp() string {
	addrList, _ := net.InterfaceAddrs()

	for _, addr := range addrList {
		// 检查ip地址判断是否回环地址
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}

	return ""
}

func HostIP() (net.IP, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	bestScore := -1
	var bestIP net.IP
	// Select the highest scoring IP as the best IP.
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			// Skip this interface if there is an error.
			continue
		}

		for _, addr := range addrs {
			score, ip := scoreAddr(iface, addr)
			if score > bestScore {
				bestScore = score
				bestIP = ip
			}
		}
	}

	if bestScore == -1 {
		return nil, errors.New("no addresses to listen on")
	}

	return bestIP, nil
}

func scoreAddr(iface net.Interface, addr net.Addr) (int, net.IP) {
	var ip net.IP
	if netAddr, ok := addr.(*net.IPNet); ok {
		ip = netAddr.IP
	} else if netIP, ok := addr.(*net.IPAddr); ok {
		ip = netIP.IP
	} else {
		return -1, nil
	}

	var score int
	if ip.To4() != nil {
		score += 300
	}
	if iface.Flags&net.FlagLoopback == 0 && !ip.IsLoopback() {
		score += 100
		if iface.Flags&net.FlagUp != 0 {
			score += 100
		}
	}
	return score, ip
}

func InStringArray(item string, items []string) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}

func InInt64Array(item int64, items []int64) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}

func Random(min, max int64) int64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int63n(max-min+1) + min
}

// SplitArray 分组
func SplitArray(arr []string, num int64) (splits [][]string) {
	var ltSlices = make([][]string, 0)
	max := int64(len(arr))
	if max < num || num <= 0 {
		ltSlices[0] = arr
		return ltSlices
	}
	var quantity int64
	if max%num == 0 {
		quantity = max / num
	} else {
		quantity = (max / num) + 1
	}
	var start, end, i int64
	for i = 1; i <= num; i++ {
		end = i * quantity
		if i != num {
			ltSlices = append(ltSlices, arr[start:end])
		} else {
			ltSlices = append(ltSlices, arr[start:])
		}
		start = i * quantity
	}
	return ltSlices
}

func SplitIntArray(arr []int64, num int64) [][]int64 {
	max := int64(len(arr))
	//判断数组大小是否小于等于指定分割大小的值，是则把原数组放入二维数组返回
	if max <= num {
		return [][]int64{arr}
	}
	//获取应该数组分割为多少份
	var quantity int64
	if max%num == 0 {
		quantity = max / num
	} else {
		quantity = (max / num) + 1
	}
	//声明分割好的二维数组
	var segments = make([][]int64, 0)
	//声明分割数组的截止下标
	var start, end, i int64
	for i = 1; i <= num; i++ {
		end = i * quantity
		if i != num {
			segments = append(segments, arr[start:end])
		} else {
			segments = append(segments, arr[start:])
		}
		start = i * quantity
	}
	return segments
}

// ExceptArray 从src中去除except
func ExceptArray(src []int64, except []int64) []int64 {
	newArray := make([]int64, 0)

	excMap := make(map[int64]bool)
	for _, value := range except {
		excMap[value] = true
	}

	for _, value := range src {
		if _, ok := excMap[value]; ok {
			continue
		}

		newArray = append(newArray, value)
	}

	return newArray
}

//int64 数组去重
func UniqueInt64Slice(src []int64) []int64 {
	result := make([]int64, 0)
	tempMap := map[int64]byte{}
	for _, e := range src {
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l {
			result = append(result, e)
		}
	}
	return result
}

// Stack 调用栈
func Stack() []byte {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return buf[:n]
}

func Md5(s string) string {
	data := []byte(s)
	return fmt.Sprintf("%x", md5.Sum(data))
}

func Data2Md5(data interface{}) string {
	bytesVal, _ := json.Marshal(data)
	return fmt.Sprintf("%x", md5.Sum(bytesVal))
}

// GenRandString 生成指定长度随机字符串
// 0 - 数字
// 1 - 小写字母
// 2 - 大写字母
// 3 - 数字、小写、大写字母
func GenRandString(size int, kind int) string {
	randKind := kind
	kinds := [][]int{{10, 48}, {26, 97}, {26, 65}}
	result := make([]byte, size)

	isAll := kind > 2 || kind < 0

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < size; i++ {
		if isAll { // random randKind
			randKind = rand.Intn(3)
		}
		scope, base := kinds[randKind][0], kinds[randKind][1]
		result[i] = uint8(base + rand.Intn(scope))
	}

	return string(result)
}

// GenFaceInitSignSHA 字符串编码SHA1
func GenFaceInitSignSHA(src string) string {
	h := sha1.New()
	h.Write([]byte(src))
	bs := h.Sum(nil)
	return fmt.Sprintf("%X", bs)
}

// StringToInt64Slice 逗号分割string转[]int64
func StringToInt64Slice(src string) []int64 {
	var result []int64
	stringList := strings.Split(src, ",")
	for _, v := range stringList {
		vInt, _ := strconv.ParseInt(v, 10, 64)
		result = append(result, vInt)
	}
	return result
}

func StringSlice2InterfaceSlice(src []string) []interface{} {
	result := make([]interface{}, 0)
	for _, v := range src {
		result = append(result, v)
	}
	return result
}

func RemoveDuplicateElementString(array []string) []string {
	result := make([]string, 0, len(array))
	temp := map[string]struct{}{}
	for _, item := range array {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// Int64Slice2String []int64 to []string
func Int64Slice2String(src []int64) string {
	var res []string
	for _, v := range src {
		res = append(res, fmt.Sprintf("%d", v))
	}
	return strings.Join(res, ",")
}

func StringSlice2Int64Slice(src []string) []int64 {
	var res []int64
	for _, v := range src {
		num, _ := strconv.ParseInt(v, 10, 64)
		res = append(res, num)
	}
	return res
}

// ExceptArrayByInt64 从src中去除except
func ExceptArrayByInt64(src []int64, except int64) []int64 {
	newArray := make([]int64, 0)
	for _, value := range src {
		if except == value {
			continue
		}
		newArray = append(newArray, value)
	}
	return newArray
}

func RandomInt64Slice(src []int64) (int64, error) {
	if len(src) <= 0 {
		return 0, errors.New("slice is empty")
	}
	rand.Seed(time.Now().UnixNano())
	index := rand.Int63n((int64(len(src)) - 1) + 1)
	return src[index], nil
}

func RandomStringSlice(src []string) (string, error) {
	if len(src) <= 0 {
		return "", errors.New("slice is empty")
	}
	rand.Seed(time.Now().UnixNano())
	index := rand.Int63n((int64(len(src)) - 1) + 1)
	return src[index], nil
}

// Decimal 保留两位小数
func Decimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}

func StringPointerVal(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}

func Int64PointerVal(num *int64) int64 {
	if num == nil {
		return 0
	}
	return *num
}

func TimePointerVal(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

// VerifyEmailFormat email verify
func VerifyEmailFormat(email string) bool {
	//pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*` //匹配电子邮箱
	pattern := `^[0-9a-z][_.0-9a-z-]{0,31}@([0-9a-z][0-9a-z-]{0,30}[0-9a-z]\.){1,4}[a-z]{2,4}$`

	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

// VerifyMobileFormat mobile verify
func VerifyMobileFormat(mobileNum string) bool {
	regular := "^((13[0-9])|(14[5,7])|(15[0-3,5-9])|(17[0,3,5-8])|(18[0-9])|166|198|199|(147))\\d{8}$"

	reg := regexp.MustCompile(regular)
	return reg.MatchString(mobileNum)
}

func GetHost() string {
	var host = os.Getenv("FINAL_HOST")
	if len(host) <= 0 {
		host = GetIntranetIp()
	}
	return host
}

func String2Float(str string) (float64, error) {
	// 价格类型转换
	var num float64
	var err error
	if len(str) > 0 {
		num, err = strconv.ParseFloat(str, 64)
	}
	return num, err
}

func IsNum(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func MysqlReplace(source string) string {
	var escapeReplacer = strings.NewReplacer(`%`, `\%`, `_`, `\_`, `\`, `\\`)
	return escapeReplacer.Replace(source)
}

// RunCmdPipe 使用管道执行命令
// 调用示例：
//  ping := exec.Command("ping", "10.255.240.129", "-c", "1")
//	cmdArgs := []*exec.Cmd{
//		exec.Command("grep", "from"),
//		exec.Command("awk", "{print$7}"),
//		exec.Command("awk", "-F", "=", "{print$2}"),
//	}
//	RunCmdPipe(ping, cmdArgs...)
func RunCmdPipe(first *exec.Cmd, args ...*exec.Cmd) (string, error) {
	var current = first

	var (
		buffer bytes.Buffer
		err    error
	)

	for k, v := range args {
		args[k].Stdin, err = current.StdoutPipe()
		if err != nil {
			return "", err
		}
		if k == len(args)-1 {
			args[k].Stdout = &buffer
		}
		current = v
	}

	for _, v := range args {
		err = v.Start()
		if err != nil {
			return "", err
		}
	}

	err = first.Run()
	if err != nil {
		return "", err
	}

	for _, v := range args {
		err = v.Wait()
		if err != nil {
			return "", err
		}
	}
	return buffer.String(), err
}
