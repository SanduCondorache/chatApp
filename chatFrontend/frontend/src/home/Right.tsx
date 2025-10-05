import { useState } from "react";
import { SendMsgBetweenUsers as SendMsg } from "../../wailsjs/go/main/App.js";

type RightViewProps = {
    selected: string;
    sender: string;
    onlineMap: Record<string, boolean>;
};

export function RightView({ selected, sender, onlineMap }: RightViewProps) {
    const [msg, setMsg] = useState("");
    const [messages, setMessages] = useState<string[]>([]);

    const handleMsgInsert = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        if (!selected) return;

        try {
            const result = await SendMsg(sender, selected, msg);
            if (result === "message_sent") {
                setMessages(prev => [...prev, msg]);
                setMsg("");
            }
            else console.log("Message recv failed");
        } catch (err: any) {
            console.log(err.toString());
        }
    };


    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
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
            <div className="messages">
                {messages.map((m, i) => (
                    <p key={i}>{m}</p>
                ))}
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
