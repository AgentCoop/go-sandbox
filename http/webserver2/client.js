
var fetch = require("cross-fetch")
require ("./build/gen/api/type/chat_pb")

let createRoomReq = new proto.chat.Room.CreateRequest

createRoomReq
   .setName("My Room")
   .setDesc("test")

fetch("https://sandbox.io:8080", {
    method: "POST",
    body: createRoomReq.serializeBinary()
}).then((response) => {
    response.text().then(t => {
        console.log(t)
    })
})