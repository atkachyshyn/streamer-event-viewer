import { useRouter } from 'next/router'
import { TwitchEmbed } from 'react-twitch-embed' 

export default function Streamer() {
    const router = useRouter()
    const { name } = router.query

    console.log(name, router.query)

    return <React.Fragment>
        <TwitchEmbed
            channel={name}
            id={name}
            theme="dark"
            muted
            onVideoPause={() => console.log(':(')}
        />
    </React.Fragment>
}