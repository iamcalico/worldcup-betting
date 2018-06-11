package main

import (
	"database/sql"
	"os"
)

var (
	db *sql.DB

	countryMap = map[string]int{
		"待定":    0,
		"俄罗斯":   1,
		"沙特阿拉伯": 2,
		"埃及":    3,
		"乌拉圭":   4,
		"摩洛哥":   5,
		"伊朗":    6,
		"葡萄牙":   7,
		"西班牙":   8,
		"法国":    9,
		"澳大利亚":  10,
		"阿根廷":   11,
		"冰岛":    12,
		"秘鲁":    13,
		"丹麦":    14,
		"克罗地亚":  15,
		"尼日利亚":  16,
		"哥斯达黎加": 17,
		"塞尔维亚":  18,
		"德国":    19,
		"墨西哥":   20,
		"巴西":    21,
		"瑞士":    22,
		"瑞典":    23,
		"韩国":    24,
		"比利时":   25,
		"巴拿马":   26,
		"突尼斯":   27,
		"英格兰":   28,
		"哥伦比亚":  29,
		"日本":    30,
		"波兰":    31,
		"塞内加尔":  32,
	}
	CountryInfoList = []CountryInfo{
		{0, "待定", "unknown"},
		{1, "俄罗斯", "http://flags.fmcdn.net/data/flags/w1160/ru.png"},
		{2, "沙特阿拉伯", "http://flags.fmcdn.net/data/flags/w1160/sa.png"},
		{3, "埃及", "http://flags.fmcdn.net/data/flags/w1160/eg.png"},
		{4, "乌拉圭", "http://flags.fmcdn.net/data/flags/w1160/uy.png"},
		{5, "摩洛哥", "http://flags.fmcdn.net/data/flags/w1160/ma.png"},
		{6, "伊朗", "http://flags.fmcdn.net/data/flags/w1160/ir.png"},
		{7, "葡萄牙", "http://flags.fmcdn.net/data/flags/w1160/pt.png"},
		{8, "西班牙", "http://flags.fmcdn.net/data/flags/w1160/es.png"},
		{9, "法国", "http://flags.fmcdn.net/data/flags/w1160/fr.png"},
		{10, "澳大利亚", "http://flags.fmcdn.net/data/flags/w1160/au.png"},
		{11, "阿根廷", "http://flags.fmcdn.net/data/flags/w1160/ar.png"},
		{12, "冰岛", "http://flags.fmcdn.net/data/flags/w1160/is.png"},
		{13, "秘鲁", "http://flags.fmcdn.net/data/flags/w1160/pe.png"},
		{14, "丹麦", "http://flags.fmcdn.net/data/flags/w1160/dk.png"},
		{15, "克罗地亚", "http://flags.fmcdn.net/data/flags/w1160/hr.png"},
		{16, "尼日利亚", "http://flags.fmcdn.net/data/flags/w1160/ng.png"},
		{17, "哥斯达黎加", "http://flags.fmcdn.net/data/flags/w1160/cr.png"},
		{18, "塞尔维亚", "http://flags.fmcdn.net/data/flags/w1160/rs.png"},
		{19, "德国", "http://flags.fmcdn.net/data/flags/w1160/de.png"},
		{20, "墨西哥", "http://flags.fmcdn.net/data/flags/w1160/mx.png"},
		{21, "巴西", "http://flags.fmcdn.net/data/flags/w1160/br.png"},
		{22, "瑞士", "http://flags.fmcdn.net/data/flags/w1160/ch.png"},
		{23, "瑞典", "http://flags.fmcdn.net/data/flags/w1160/se.png"},
		{24, "韩国", "http://flags.fmcdn.net/data/flags/w1160/kr.png"},
		{25, "比利时", "http://flags.fmcdn.net/data/flags/w1160/be.png"},
		{26, "巴拿马", "http://flags.fmcdn.net/data/flags/w1160/pa.png"},
		{27, "突尼斯", "http://flags.fmcdn.net/data/flags/w1160/tn.png"},
		{28, "英格兰", "https://upload.wikimedia.org/wikipedia/en/thumb/b/be/Flag_of_England.svg/800px-Flag_of_England.svg.png"},
		{29, "哥伦比亚", "http://flags.fmcdn.net/data/flags/w1160/co.png"},
		{30, "日本", "http://flags.fmcdn.net/data/flags/w1160/jp.png"},
		{31, "波兰", "http://flags.fmcdn.net/data/flags/w1160/pl.png"},
		{32, "塞内加尔", "http://flags.fmcdn.net/data/flags/w1160/sn.png"},
	}

	config           Config
	GdisplayFileList = []string{}

	timiUserWhiteList = make(map[string]string)
	timiNewUsers      *os.File
)
