
module.exports = async function (context) {
    console.log(context.request.body);
    console.log("z-custom-name: " + context.request.headers['z-custom-name']);
    console.log("x-kubefaas-function-name: " + context.request.headers['x-kubefaas-function-name']);
    let obj = context.request.body;
    let headers = context.request.headers;
    return {
        status: 200,
	headers: headers,
        body: obj
    };
}
