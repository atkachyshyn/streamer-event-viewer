// Next.js API route support: https://nextjs.org/docs/api-routes/introduction

export default async (req, res) => {
  try {
    const { query: { name }, headers: { cookie } } = req

    console.log("http://localhost:8080/subscribe", name, req)

    let header = new Headers({
      // 'Access-Control-Allow-Origin':'*',
      'Content-Type': 'application/json',
      'Cookie': cookie || ''
    })

    console.log("header", header)

    const response = await fetch('http://localhost:8080/subscribe', {
      method: 'post',
      headers: header,
      mode: "cors",
      body: JSON.stringify({
        streamer: name
      })
    })

    console.log(response.headers)
  } catch(err) {
    alert(err)
  }

  res.statusCode = 200
  res.json({ name: 'John Doe' })
}
