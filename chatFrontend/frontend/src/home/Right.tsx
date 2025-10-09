import { useEffect, useState } from "react";
import { SendMsgBetweenUsers as SendMsg } from "../../wailsjs/go/main/App.js";
import { EventsOn, EventsOff } from "../../wailsjs/runtime/runtime.js";
import { ChatMessage } from "../types/ChatMessages.js";
import { MessageHist } from "../types/MessageHist.js";

type RightViewProps = {
    selected: string;
    sender: string;
    mess: MessageHist[];
    onlineMap: Record<string, boolean>;
};

export function RightView({ selected, sender, mess, onlineMap }: RightViewProps) {
    const [msg, setMsg] = useState("");
    const [messages, setMessages] = useState<MessageHist[]>(mess);

    useEffect(() => {
        setMessages(mess);
    }, [mess]);

    useEffect(() => {
        const handler = (payload: string) => {
            const msg = JSON.parse(payload) as ChatMessage;
            console.log("chat event received", msg.msg, sender)
            if (msg.recv_id !== sender) return;
            setMessages(prev => [
                ...prev,
                { direction: "received", content: msg.msg, time: new Date(msg.created_at).toString() }
            ]);
        };

        EventsOn("chat:received", handler);
        return () => EventsOff("chat:received");
    }, [sender]);


    const handleMsgInsert = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        if (!selected) return;

        try {
            const result = await SendMsg(sender, selected, msg);
            if (result === "message_sent") {
                let temp: MessageHist;
                temp = {
                    direction: "sent",
                    content: msg,
                    time: new Date().toString()
                }

                setMessages(prev => [...prev, temp]);
                setMsg("");
            }
        } catch (err: any) {
            console.log(err.toString());
        }
    };


    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        if (!selected) return;
        setMsg(e.target.value);
    };

    return (
        <div className="split-pane right-pane">
            <div className="chat-header">
                {selected ? (
                    <div className="chat-top">
                        <div className="avatar">{selected[0].toUpperCase()}</div>
                        <div className="chat-block">
                            <h2 className="chat-title">{selected}</h2>
                            <p className="chat-subtitle">{onlineMap[selected] ? "Online" : "Offline"}</p>
                        </div>
                    </div>
                ) : (
                    <h2 className="chat-title">Right Pane</h2>
                )}
            </div>
            <div className="messages chat-container1">
                {messages.map((m, i) => {
                    if (m.direction === "sent") {
                        return <div className="message sent-messages" key={i}>{m.content}</div>;
                    } else {
                        return <div className="message received-messages" key={i}>{m.content}</div>;
                    }
                })}
            </div>
            <div className="input-bar">
                <form onSubmit={handleMsgInsert}>
                    <input
                        type="text"
                        placeholder="Type a message..."
                        value={msg}
                        onChange={handleChange}
                    />
                    <button type="submit" style={{ display: "none" }} />
                </form>
            </div>
        </div>
    );
}
