import { useState } from "react";
import { SendMsgBetweenUsers as SendMsg } from "../../wailsjs/go/main/App.js";

type RightViewProps = {
    selected: string;
    sender: string;
}

export function RightView({ selected, sender }: RightViewProps) {
    const [msg, setmsg] = useState("");

    const handleMsgInsert = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();

        try {
            const result = await SendMsg(sender, selected, msg);
            if (result == "ok") {
                console.log("Message sent" + result);
            } else {
                console.log("Message recv");
            }
        } catch (err: any) {
            console.log(err.toString());

        }
    }

    return (
        <div className="split-pane right-pane">
            <div className="messages">
                {selected ? (
                    <h2>{selected}</h2>
                ) : (
                    <h2>Right Pane</h2>
                )}
            </div>

            <div className="input-bar">
                <form onSubmit={handleMsgInsert}>
                    <input
                        type="text"
                        placeholder="Type a message..."
                        value={msg}
                        onChange={(e) => {
                            setmsg(e.target.value);
                        }}
                    />
                    <button type="submit" style={{ display: "none" }} />
                </form>
            </div>
        </div>
    )
}

