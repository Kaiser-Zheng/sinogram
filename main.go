package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"strings"
)

const (
	defaultDictFile = "dictionary.md"
	base64Charset   = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	minDictChars    = 256  // Minimum characters for basic functionality
	maxPairs        = 4096 // 64 * 64 possible base64 pairs
)

// Codec handles encoding/decoding between base64 pairs and Chinese characters
type Codec struct {
	pairToRune map[string]rune
	runeToPair map[rune]string
}

func NewCodec() *Codec {
	return &Codec{
		pairToRune: make(map[string]rune),
		runeToPair: make(map[rune]string),
	}
}

// LoadDictionary builds the character mapping from a Chinese text file
func (c *Codec) LoadDictionary(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read dictionary: %w", err)
	}

	uniqueChars := extractChineseCharacters(string(content))

	if len(uniqueChars) < minDictChars {
		return fmt.Errorf("insufficient Chinese characters (found: %d, need: %d+)",
			len(uniqueChars), minDictChars)
	}

	c.buildMapping(uniqueChars)
	c.printStats(len(uniqueChars))

	return nil
}

// extractChineseCharacters collects unique Chinese characters from text
func extractChineseCharacters(text string) []rune {
	seen := make(map[rune]bool)
	var chars []rune

	for _, r := range text {
		if !seen[r] && isChineseChar(r) {
			chars = append(chars, r)
			seen[r] = true
		}
	}

	return chars
}

// isChineseChar checks if rune is in CJK Unicode ranges
func isChineseChar(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
		(r >= 0x3400 && r <= 0x4DBF) || // CJK Extension A
		(r >= 0x20000 && r <= 0x2A6DF) // CJK Extension B
}

// buildMapping creates bidirectional mapping between base64 pairs and Chinese chars
func (c *Codec) buildMapping(chars []rune) {
	idx := 0

	for i := 0; i < 64 && idx < len(chars); i++ {
		for j := 0; j < 64 && idx < len(chars); j++ {
			pair := string([]byte{base64Charset[i], base64Charset[j]})
			char := chars[idx]

			c.pairToRune[pair] = char
			c.runeToPair[char] = pair
			idx++
		}
	}
}

func (c *Codec) printStats(totalChars int) {
	coverage := len(c.pairToRune)
	fmt.Printf("Dictionary loaded: %d unique Chinese characters\n", totalChars)
	fmt.Printf("Coverage: %d/%d pairs (%.1f%%)\n",
		coverage, maxPairs, float64(coverage)/maxPairs*100)
}

// Encode converts a file to Chinese character representation
func (c *Codec) Encode(inputPath, outputPath string, useBase64 bool) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	encoded := c.encodeData(data, useBase64)

	if err := os.WriteFile(outputPath, []byte(encoded), 0644); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	c.printEncodeStats(len(data), len(encoded), useBase64)
	return nil
}

func (c *Codec) encodeData(data []byte, useBase64 bool) string {
	var text string
	if useBase64 {
		text = base64.StdEncoding.EncodeToString(data)
	} else {
		text = string(data)
	}

	// Pad to even length
	if len(text)%2 != 0 {
		text += "="
	}

	var result strings.Builder
	unmapped := 0

	for i := 0; i < len(text)-1; i += 2 {
		pair := text[i : i+2]

		// Only map valid base64 character pairs
		if c.isValidBase64Pair(pair) {
			if char, ok := c.pairToRune[pair]; ok {
				result.WriteRune(char)
				continue
			}
			unmapped++
		}

		// Keep unmapped or invalid pairs as-is
		result.WriteString(pair)
	}

	if unmapped > 0 {
		fmt.Printf("Warning: %d pairs not in dictionary\n", unmapped)
	}

	return result.String()
}

func (c *Codec) isValidBase64Pair(pair string) bool {
	return len(pair) == 2 &&
		strings.IndexByte(base64Charset, pair[0]) != -1 &&
		strings.IndexByte(base64Charset, pair[1]) != -1
}

func (c *Codec) printEncodeStats(inputSize, outputSize int, useBase64 bool) {
	fmt.Printf("Original size: %d bytes\n", inputSize)
	if useBase64 {
		b64Size := base64.StdEncoding.EncodedLen(inputSize)
		fmt.Printf("Base64 size: %d bytes\n", b64Size)
	}
	fmt.Printf("Encoded size: %d bytes\n", outputSize)
	fmt.Printf("Encoding complete: output saved\n")
}

// Decode converts Chinese character representation back to original data
func (c *Codec) Decode(inputPath, outputPath string, useBase64 bool) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	decoded, err := c.decodeData(string(data), useBase64)
	if err != nil {
		return fmt.Errorf("decode failed: %w", err)
	}

	if err := os.WriteFile(outputPath, decoded, 0644); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	fmt.Printf("Decoding complete: %d bytes written\n", len(decoded))
	return nil
}

func (c *Codec) decodeData(text string, useBase64 bool) ([]byte, error) {
	var base64Text strings.Builder

	// Convert Chinese characters back to base64 pairs
	for _, r := range text {
		if pair, ok := c.runeToPair[r]; ok {
			base64Text.WriteString(pair)
		} else {
			// Character not in mapping, keep as-is (handles padding, etc.)
			base64Text.WriteRune(r)
		}
	}

	base64Str := base64Text.String()

	if useBase64 {
		return base64.StdEncoding.DecodeString(base64Str)
	}

	return []byte(base64Str), nil
}

// generateSampleDictionary creates a default dictionary file
func generateSampleDictionary(filename string) error {
	sample := `第一回甄士隱夢幻識通靈賈雨村風塵懷閨秀作者自云因曾歷過番之後故將真事去而借說撰此石頭記書也曰但中所何又是哉今碌無成忽念及當日有女子細推了覺其行止見皆出於我上堂鬚眉不若彼裙釵實愧則餘悔益大可奈時欲已往賴天恩下承祖德錦衣紈絝飫甘饜美肥背父母教育負師兄規訓至半生潦倒罪編述以告普人固能免然閣本萬肖護短併使泯滅雖茆椽蓬牖瓦灶繩床晨月夕階柳庭花亦未傷襟筆墨學文爲用假語言敷演段來悅耳目乃題綱正義開卷即知意原友情並非怨世駡矣涉態得叙旨閱切詩浮着甚苦奔忙盛席華筵終散場悲喜千般同渺古盡荒唐謾紅袖啼痕重更痴抱恨長字看血十年辛尋常列位官你道從起根由近諳深趣味待在注明方聞惑媧氏煉補山稽崖高經二丈四頑三六五百零塊皇只單剩便棄青埂峰誰煅性眾俱獨己材堪入選遂嗟夜號慚悼俄僧遠骨骼凡豐神迥异笑坐邊談快論先些雲霧海僊玄到榮富貴聽打動心想要間享這粗蠢口吐向那弟物禮適耀繁慕質卻稍況仙形體定品必濟利如蒙發點慈攜帶溫柔鄉裡受幾永佩洪劫忘畢齊憨善樂依恃足好多魔八個緊相連屬瞬息极換究竟境歸空的熾進話复求再強制歎靜極思數既們莫并奇處踮腳罷施佛法助還案否感謝咒符展術登變鮮瑩洁玉且縮扇墜小拿托掌体寶沒須鐫妙昌隆邦簪纓族地安身業禁問件携望乞示白飄投舍訪跡分就茫離合歡炎涼面首偈與蒼枉許係前倩寄傳落胎親陳家瑣閒詞全備或适解悶朝代紀輿國反失考据寫賢忠理廷治俗政樣才微班姑蔡縱抄恐愛呢答太耶漢等添綴難野史蹈轍套新別致取拘市井少特訕謗君貶妻奸淫凶惡胜种穢污臭屠毒坏佳部共濫滿紙潘建西兩艷賦擬男名姓旁撥亂劇丑鬟婢乎逐悉矛盾睹敢似委消愁破歪熟噴飯供酒興衰際遇追蹤躡加穿鑿徒為貧食累怀貪戀色貨工夫願稱檢讀他醉飽臥避把玩豈省壽命筋力比謀虛妄舌害腿令眼胡牽扯淑娘舊稿忖晌遍指責佞誅邪罵仁臣良孝倫關功頌眷窮錄邀約私訂偷盟毫干尾悟易改吳樓東魯孔梅溪鑒曹雪芹軒披載增刪次纂章金陵絕酸淚都脂硯齋甲戌評仍按陷南隅蘇城閶門最流外里街內清巷廟窄狹呼葫蘆住宦費嫡封稟恬淡每觀修竹酌吟膝兒乳喚英蓮歲夏晝房手倦拋伏几憩朦朧睡辨廂放現公該結冤尚趁機會夾孽造罕河岸畔絳珠草株赤瑕宮瑛侍露灌溉始久延精滋養脫木僅游饑蜜果膳渴飲水湯酬報郁纏綿恰偶乘平緣警挂償惠勾陪碎膩概篇總香竊暗泄鬼愚度吾交割楚完猶集隨系請濁洞洗諦沉預跳火坑遞接奪牌坊幅對聯跟舉步聲霹靂崩叫睛烈芭蕉冉奶走越粉妝琢乖伸鬥耍熱鬧癩跣跛瘋癲揮霍哭主運爹睬耐煩撤句慣嬌菱澌防節元宵煙豫各幹營北邙銷影試晚隔壁居儒化表飛州仕末宗基喪京整淹蹇暫賣老倚佇引聊送童獻茶嚴爺拜慌恕誑駕略讓客候妨廳翻弄籍窗嗽丫擷儀容姿呆猛抬敝巾服窘腰圓厚闊兼劍星直鼻權腮轉雄壯襤褸什麼幫周疑怪困狂巨留早秋宴另具顧刻值占律卜頻斂額儔蟾光逢搔匵价奩淺誕謂團尊旅寂寥納辭拂院臾設杯盤肴款斟漫漸濃觥限斝簫管戶弦歌輪彩凝輝愈豪乾七寓晴欄捧仰騰兆履霓賀斗充沽囊路措突宜速春闈戰置謬銀冬九黃期買舟晤收介吃竿醒昨荐謁和鼓達黑陰倏霄啟社燈檻急逃婦妥找音響旦死病孺构疾醫療炸油鍋逸燒篱抵條焰軍民救勢熄怜片礫惟跌商議田庄偏旱鼠盜蜂搶狗兵剿捕折岳肅貫務農殷婿狼狽幸薄計哄賺朽屋稼穡勉支持活懶驚唬忿痛積暮攻景巧拄拐杖掙挫麻屣鶉曉冢堆聚閉孫順迎算宿慧徹陋室笏枯楊舞蛛絲雕梁綠紗糊鬢霜土隴帳底鴛鴦箱丐嘆保擇膏粱嫌帽鎖枷扛憐襖寒紫蟒唱認嫁裳拍肩褡褳烘信遣討靠僕針線喝任牢轎烏猩袍府怔象丟歇嚷差瞪禍`

	return os.WriteFile(filename, []byte(sample), 0644)
}

func main() {
	// Command-line flags
	encodeFile := flag.String("e", "", "Encode: specify input file")
	decodeFile := flag.String("d", "", "Decode: specify input file")
	dictFile := flag.String("dict", defaultDictFile, "Dictionary file path")
	outputFile := flag.String("o", "", "Output file name")
	useBase64 := flag.Bool("b64", true, "Use base64 encoding (default: true)")
	genDict := flag.Bool("gen-dict", false, "Generate sample dictionary")

	flag.Parse()

	// Generate dictionary if requested
	if *genDict {
		if err := generateSampleDictionary(defaultDictFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Sample dictionary generated: %s\n", defaultDictFile)
		return
	}

	// Initialize codec and load dictionary
	codec := NewCodec()
	if err := codec.LoadDictionary(*dictFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Handle encoding
	if *encodeFile != "" {
		output := *outputFile
		if output == "" {
			output = *encodeFile + ".encoded"
		}

		if err := codec.Encode(*encodeFile, output, *useBase64); err != nil {
			fmt.Fprintf(os.Stderr, "Encoding error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle decoding
	if *decodeFile != "" {
		output := *outputFile
		if output == "" {
			output = *decodeFile + ".decoded"
		}

		if err := codec.Decode(*decodeFile, output, *useBase64); err != nil {
			fmt.Fprintf(os.Stderr, "Decoding error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// No action specified
	flag.Usage()
}
