import Head from 'next/head'
import Link from 'next/link'
import { useState, useEffect } from 'react'
import useSWR from 'swr'

const fetcher = async (url) => {
  const res = await fetch(url)
  const data = await res.json()

  if (res.status !== 200) {
    throw new Error(data.message)
  }
  return data
}

export default function Home() {
  const name = 'AhriNyan'

  const [input, setInput] = useState('')
  const [shouldSubscribe, setShouldSubscribe] = React.useState(false);

  const { data } = useSWR(() => shouldSubscribe ? `/api/subscribe/${input}` : null, fetcher);

  function handleClick() {
    console.log("shouldSubscribe: ", shouldSubscribe)
    setShouldSubscribe(true);
  }

  console.log(data)

  // const subscribe = async (e) => {
  //   e.preventDefault()
  //   try {
  //     const { data, error } = useSWR(
  //       () => `/api/subscribe/${input}`,
  //       fetcher
  //     )

  //     console.log(data)
  //   } catch(err) {
  //     alert(err)
  //   }
  // }

  return (
    <div className="container">
      <Head>
        <title>SEV</title>
      </Head>

      <main>
        <h1 className="title">
          <span>StreamerEventViewer</span>
        </h1>

        <p className="description">
          <a href="http://localhost:8080/login">
            Login with your <code>Twitch</code> account
          </a>
        </p>

        <div className="grid">
          <Link
            href="/streamer/[name]"
            as={`/streamer/AhriNyan`}
          >
            <a>
              <h3>Watch your favourite streamer</h3>
              <p>
                Watch live stream of your favourite streamer with chat and recent events.
              </p>
            </a>
          </Link>
          <div className='flex'>
            <input className='bg-gray-200 shadow-inner rounded-l p-2 flex-1' id='name' type='input' aria-label='streamer name' placeholder='Enter your favourite streamer' value={input} onChange={e => setInput(e.target.value)} />
            <button disable={shouldSubscribe} onClick={handleClick}>Fetch</button>
          </div>
        </div>
      </main>

      <footer>
      </footer>

      <style jsx>{`
        .container {
          min-height: 100vh;
          padding: 0 0.5rem;
          display: flex;
          flex-direction: column;
          justify-content: center;
          align-items: center;
        }

        main {
          padding: 5rem 0;
          flex: 1;
          display: flex;
          flex-direction: column;
          justify-content: center;
          align-items: center;
        }

        footer {
          width: 100%;
          height: 100px;
          border-top: 1px solid #eaeaea;
          display: flex;
          justify-content: center;
          align-items: center;
        }

        footer img {
          margin-left: 0.5rem;
        }

        footer a {
          display: flex;
          justify-content: center;
          align-items: center;
        }

        a {
          color: inherit;
          text-decoration: none;
        }

        .title a {
          color: #0070f3;
          text-decoration: none;
        }

        .title a:hover,
        .title a:focus,
        .title a:active {
          text-decoration: underline;
        }

        .title {
          margin: 0;
          line-height: 1.15;
          font-size: 4rem;
        }

        .title,
        .description {
          text-align: center;
        }

        .description {
          line-height: 1.5;
          font-size: 1.5rem;
        }

        code {
          background: #fafafa;
          border-radius: 5px;
          padding: 0.75rem;
          font-size: 1.1rem;
          font-family: Menlo, Monaco, Lucida Console, Liberation Mono,
            DejaVu Sans Mono, Bitstream Vera Sans Mono, Courier New, monospace;
        }

        .grid {
          display: flex;
          align-items: center;
          justify-content: center;
          flex-wrap: wrap;

          max-width: 800px;
          margin-top: 3rem;
        }

        .card {
          margin: 1rem;
          flex-basis: 45%;
          padding: 1.5rem;
          text-align: left;
          color: inherit;
          text-decoration: none;
          border: 1px solid #eaeaea;
          border-radius: 10px;
          transition: color 0.15s ease, border-color 0.15s ease;
        }

        .card:hover,
        .card:focus,
        .card:active {
          color: #0070f3;
          border-color: #0070f3;
        }

        .card h3 {
          margin: 0 0 1rem 0;
          font-size: 1.5rem;
        }

        .card p {
          margin: 0;
          font-size: 1.25rem;
          line-height: 1.5;
        }

        .logo {
          height: 1em;
        }

        @media (max-width: 600px) {
          .grid {
            width: 100%;
            flex-direction: column;
          }
        }
      `}</style>

      <style jsx global>{`
        html,
        body {
          padding: 0;
          margin: 0;
          font-family: -apple-system, BlinkMacSystemFont, Segoe UI, Roboto,
            Oxygen, Ubuntu, Cantarell, Fira Sans, Droid Sans, Helvetica Neue,
            sans-serif;
        }

        * {
          box-sizing: border-box;
        }
      `}</style>
    </div>
  )
}
