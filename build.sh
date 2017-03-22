#!/bin/bash

mypath=git.oschina.net/jkkkls/goxiang/


go test ${mypath}GxMessage
go build ${mypath}GxMessage
# go install ${mypath}GxMessage

go test ${mypath}GxMisc
go build ${mypath}GxMisc
# go install ${mypath}GxMisc

go test ${mypath}GxNet
go build ${mypath}GxNet
# go install ${mypath}GxNet

go test ${mypath}GxStatic
go build ${mypath}GxStatic
# go install ${mypath}GxStatic

go test ${mypath}GxProto
go build ${mypath}GxProto
# go install ${mypath}GxProto

go test ${mypath}GxDict
go build ${mypath}GxDict
# go install ${mypath}GxDict

go build ${mypath}Login
go build ${mypath}Gate
go build ${mypath}Center
go build ${mypath}Client
go build ${mypath}Tool
go build ${mypath}Public
