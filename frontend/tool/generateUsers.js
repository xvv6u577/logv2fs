require('dotenv').config();
const axios = require("axios").default;
const jsonfile = require("jsonfile");

axios
	.get(process.env.MAIN_URL)
	.then((res) => {
		var users = [];
		for (var prop in res.data) {
			res.data[prop].forEach((item) => {
				users.push({
					email: item.email,
					path: prop,
					status: "plain",
					role: item.email === "caster" ? "admin" : "normal",
					password:
						item.email.length < 6 ? "mypassword" : item.email,
					uuid: item.id,
				});
			});
		}
		// 生成config.json
		jsonfile.writeFileSync("./tool/users.json", users);
		console.log("tool/users.json generated!");
	})
	.catch((err) => {
		console.log(err);
	});
