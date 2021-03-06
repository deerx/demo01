package mypkg

import (
	"fmt"
	"strconv"
	"sync"
)

// 修改房间状态以及插入使用记录
var mu1 sync.Mutex

// 加锁的方式获取
func GetroomLock(studentID string) (bool, string, string, string) {

	mu1.Lock()
	defer mu1.Unlock()
	var (
		flag    bool = false
		number  int
		timestr string
	)
	boolean, room := FindRoom()
	if boolean {
		DB.QueryRow(updateRoomSQL, "1", room.ID)
		timestr, endtime := GetTime()
		// 插入一条使用记录

		DB.QueryRow(insertOrderSQL, studentID, room.ID, "1", timestr, endtime).Scan(&number)
		flag = true
	} else {
		return false, "", "", ""
	}

	return flag, strconv.Itoa(room.ID), strconv.Itoa(number), timestr
}

//事物的方式获取
func GetRoomBeginTest(userNumber string) {
	// 开启事物
	tx, err1 := DB.Begin()
	if err1 != nil {
		fmt.Println("初始化事物失败")
		fmt.Println(err1)
		return
	}
	defer tx.Rollback()
	fmt.Println("初始化事物成功")
	var (
		returnRoom Room
		number     int
	)
	err := tx.QueryRow(findRoomSQL).Scan(&returnRoom.ID, &returnRoom.Type)
	if err != nil {
		fmt.Println("未找到房间回滚", err)
		fmt.Println("清理房间")
		DB.QueryRow("update room set type = 0")
		DB.QueryRow("update orders set type = 0")
		return
	}
	if returnRoom.ID != 0 {
		_, err = tx.Exec(updateRoomSQL, "1", returnRoom.ID)
		if err != nil {
			fmt.Println("修改房间状态失败执行回滚", err)
			return
		}
		timestr, endtime := GetTime()
		err = tx.QueryRow(insertOrderSQL, userNumber, returnRoom.ID, "1", timestr, endtime).Scan(&number)
		if err != nil {
			fmt.Println("插入log订单失败执行回滚", err)
			return
		}
	}
	tx.Commit()
	fmt.Println(userNumber, "预约成功房间为", returnRoom)
}
