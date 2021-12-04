package main

import (
	"database/sql"
	_ "encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	_"github.com/go-sql-driver/mysql"
	"net/http"
	_ "os"

)


type user struct {
	id int
	username string//用户名
	password string//密码
	protectionQ string//密保问题
	protectionA string//密保答案，需要输入username会显示protectionQ，
}

type comments struct {
	id       int
	towhom   string
	fromwhom string
	comments string
	commentof string
}

var db *sql.DB//是一个连接池


func main()  {
	err:= initDb()
	if err!=nil{
		fmt.Println(err)
	}
	r:=gin.Default()

	r.POST("/login",login)//登录
	r.POST("/signup",signup)//注册
	r.POST("/reset",reset)//用密保重置密码
	r.POST("/comment",cookie,comment)//评论

	r.Run(":2025")
}

func initDb()(err error){
	 db,err =sql.Open("mysql","root:@tcp(127.0.0.1:3306)/login")
	 if err!=nil{
		 fmt.Println(err)
		 return err
	 }

	 err =db.Ping()
	 if err!=nil{
		 fmt.Println(err)
		 return err
	 }

	 return nil
 }


func login(c *gin.Context){
	username:=c.PostForm("username")
	password:=c.PostForm("password")

	sqlstr:= "select username,password,id from user where username = ?;"//读取账户信息，同时判断是否存在该账户
	var u user
	err:=db.QueryRow(sqlstr,username).Scan(&u.username,&u.password,&u.id)
	if err!=nil{
		c.JSON(http.StatusOK,"没有该账户")
		return
	}
	if password==u.password{
		c.SetCookie("login_cookie", username, 3600, "/", "", false, true)
		c.JSON(http.StatusOK,gin.H{
			"登录成功，你的id为":u.id,
			"username为":u.username,
		})
	}else {
		c.JSON(http.StatusOK,"密码错误")
	}

}


func cookie(c *gin.Context){
	ck,err :=c.Cookie("login_cookie")
	if err!=nil{
		fmt.Println(err)
		c.JSON(403,"未登录")
		c.Abort()
	}else {
		c.Set("cookie",ck)
		c.Next()
	}
}

func signup (c *gin.Context){
	username:=c.PostForm("username")//用户名
	password:=c.PostForm("password")//密码
	passwordagain:=c.PostForm("passwordagain")//重复输入密码
	protectionQ:=c.PostForm("protectionQ")//密保问题
	protectionA:=c.PostForm("protectionA")//密保答案

	var u user
	err :=db.QueryRow("select username from user where username=?;",username).Scan(&u.username)//判断账号是否存在
	if err==nil {
		c.JSON(http.StatusOK,"该账号已存在")
		return
	}

	if password==passwordagain{
		_,err = db.Exec("insert into user(username,password,protectionQ,protectionA) values (?,?,?,?);",username,password,protectionQ,protectionA)//输入用户信息
		if err!=nil{
            fmt.Println(err)
		}
		c.SetCookie("login_cookie", username, 3600, "/", "", false, true)
		c.JSON(http.StatusOK,"注册成功")
	}else {
		c.JSON(http.StatusOK,"两次密码不相同")
	}
}

func reset (c *gin.Context){
	username:=c.PostForm("username")//用户名
	protectionA:=c.PostForm("protectionA")//密保答案
	newpassword:=c.PostForm("newpassword")//新密码
	newpasswordagain:=c.PostForm("newpasswordagain")//重复输入新密码

	var u user
	err := db.QueryRow("select username,protectionQ,protectionA from user where username = ?;", username).Scan(&u.username,&u.protectionQ,&u.protectionA)//检查是否有该用户，同时读取用户信息
	if err!=nil {
		fmt.Println(err)
		c.JSON(http.StatusOK,"无此账号")
		return
	}else if u.protectionQ=="" {
		c.JSON(http.StatusOK,"该账户未设置密保")
		return
	}else {
		c.JSON(http.StatusOK,gin.H{
			"你的密保问题是":u.protectionQ,
		})
		if protectionA==u.protectionA{
			if newpassword==newpasswordagain{
				_,err=db.Exec("update user set password=? where username=?;",newpassword,u.username)//更改密码
				c.JSON(http.StatusOK,"密码修改成功")
			}else {
				c.JSON(http.StatusOK,"两次密码不相同")
			}
		}else {
			c.JSON(http.StatusOK,"密保答案错误")
		}
	}
}

func comment (c *gin.Context){
	username,err:=c.Cookie("login_cookie")//从cookie中获得登录用户的用户名
	c.JSON(http.StatusOK,gin.H{
		"Hi!":username,
	})
	fmt.Println(username)
	if err!=nil {
		fmt.Println(err)
		return
	}

	towhom:=c.PostForm("to")//给谁发评论
	comment1:=c.PostForm("comment")//评论内容
	commentof:=c.PostForm("commentof")//回复于哪条内容

	sqlStr := "select comment, commentof,fromwhom from comment where towhom=?;"//遍历写给登录用户的评论
	rows, err := db.Query(sqlStr, username)
	if err != nil {
		fmt.Printf("query failed, err:%v\n", err)
		c.JSON(http.StatusOK,"无评论")
		goto sign
	}

	for rows.Next() {
		var co comments
		err := rows.Scan(&co.comments, &co.commentof, &co.fromwhom)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			return
		}
        c.JSON(http.StatusOK,gin.H{
			"来自":co.fromwhom,
			"评论了":co.comments,
			"回复于":co.commentof,
		})
	}
	rows.Close()
	sign:

	var u user
	err=db.QueryRow("select username from user where username =?;",towhom).Scan(&u.username)//检查是否有该用户
	if err!=nil{
		fmt.Println(err)
		return
	}else {
		_,err =db.Exec("insert into comment (towhom,fromwhom,comment,commentof) values (?,?,?,?);",towhom,username,comment1,commentof)//发送评论
		if err!=nil {
			fmt.Println(err)
		}
	}
}
