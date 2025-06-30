import { SIGNALING_SERVER_URL } from '@/const';
import { auth } from '@/lib/auth/auth'
import { dft } from '@/lib/dft/dft';
import { createFileRoute, redirect } from '@tanstack/react-router'
import { useEffect, useRef, useState } from 'react';

export const Route = createFileRoute('/dft/$dftID')({
    beforeLoad: async ({ params }) => {
        const user = await auth.getSession();
        if (!user) {
            throw redirect({
                to: '/auth',
            });
        }

        const dftID = params.dftID;
        const dftData = await dft.stats(dftID);
        if (!dftData) {
            throw redirect({
                to: '/dashboard',
            });
        }

        return { user, dftData };
    },
    component: Dft,
})

const configuration = {
    iceServers: [
        { urls: 'stun:stun.l.google.com:19302' }, // public STUN server
    ],
};

type ConnectionState = 'disconnected' | 'connected' | 'connecting' | 'error';

function Dft() {
    const { dftID } = Route.useParams();
    const { dftData } = Route.useRouteContext();
    const pc = useRef<RTCPeerConnection | null>(null);
    const ws = useRef<WebSocket | null>(null);

    const dataChannel = useRef<RTCDataChannel | null>(null);
    const [connectionState, setConnectionState] = useState<ConnectionState>('disconnected');
    const [messages, setMessages] = useState<string[]>([]);

    useEffect(() => {
        ws.current = new WebSocket(`${SIGNALING_SERVER_URL}?dft_id=${dftID}`);
        ws.current.onopen = () => {
            console.log('WebSocket connection established');
            setConnectionState('connected');
            startWebRTC();
        }

        ws.current.onmessage = async (event) => {
            const message = JSON.parse(event.data);
            console.log('Received message:', message);

            if (!pc.current) return;

            if (message.sdp) {
                console.log('Received SDP');
                await pc.current.setRemoteDescription(new RTCSessionDescription(message.sdp));
                if (message.sdp.type === 'offer') {
                    const answer = await pc.current.createAnswer();
                    await pc.current.setLocalDescription(answer);
                    sendSignal({ sdp: pc.current.localDescription });
                }
            } else if (message.candiate) {
                console.log('Received ICE candidate');
                try {
                    await pc.current.addIceCandidate(new RTCIceCandidate(message.candidate));
                } catch (e) {
                    console.error('Error adding received ICE candidate', e);
                }
            }
        };

        ws.current.onerror = (err) => {
            console.error('WebSocket error', err);
            setConnectionState('error');
        };

        ws.current.onclose = () => {
            console.log('WebSocket closed');
            setConnectionState('disconnected');
        };

        async function startWebRTC() {
            pc.current = new RTCPeerConnection(configuration);

            // Handle ICE candidates generated locally
            pc.current.onicecandidate = (event) => {
                if (event.candidate) {
                    sendSignal({ candidate: event.candidate });
                }
            };

            // Create data channel (for initiator)
            dataChannel.current = pc.current.createDataChannel('drift-data');

            dataChannel.current.onopen = () => {
                console.log('Data channel opened');
                setConnectionState('connected');
            };

            dataChannel.current.onmessage = (event) => {
                console.log('Data channel message:', event.data);
                setMessages((msgs) => [...msgs, event.data]);
            };

            // For the peer receiving the data channel
            pc.current.ondatachannel = (event) => {
                console.log('Data channel received');
                dataChannel.current = event.channel;
                dataChannel.current.onmessage = (e) => {
                    console.log('Data channel message:', e.data);
                    setMessages((msgs) => [...msgs, e.data]);
                };
                dataChannel.current.onopen = () => {
                    console.log('Data channel opened (receiver)');
                    setConnectionState('connected');
                };
            };

            // Create offer (for initiator)
            const offer = await pc.current.createOffer();
            await pc.current.setLocalDescription(offer);
            sendSignal({ sdp: pc.current.localDescription });
        }

        function sendSignal(message: any) {
            if (ws.current && ws.current.readyState === WebSocket.OPEN) {
                ws.current.send(JSON.stringify(message));
            }
        }

        return () => {
            pc.current?.close();
            ws.current?.close();
        };
    }, [dftID]);

    console.log(dftData);

    return (
        <div>
            <h2>Dft: {dftID}</h2>
            <p>Status: {connectionState}</p>
            <p>Drifters: {dftData.total}</p>

            <div>
                <h3>Messages received:</h3>
                <ul>
                    {messages.map((m, i) => (
                        <li key={i}>{m}</li>
                    ))}
                </ul>
            </div>
        </div>
    );
}
