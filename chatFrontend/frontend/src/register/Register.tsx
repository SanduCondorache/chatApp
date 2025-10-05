import { useNavigate } from "react-router-dom";
import "./Register.css";
import { Register as GoRegister } from "../../wailsjs/go/main/App.js"
import { useState } from "react";


export function Register() {
    const navigate = useNavigate();
    const [username, setUsername] = useState("");
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [err, setErr] = useState("");

    const handleRegister = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();

        try {
            const result = await GoRegister(username, email, password);
            if (result == "ok") {
                navigate('/home');
            } else if (result == "username_taken") {
                setErr("Username is already taken");
            }
        } catch (error: any) {
            setErr("Error:" + error.toString());

        }

    };

    return (
        <section>
            <div className="login-box">
                <form onSubmit={handleRegister}>
                    <h2>Register</h2>
                    <div className="input-box">
                        <span className='icon'></span>
                        <input type="username" value={username} onChange={e => setUsername(e.target.value)} required />
                        <label>Username</label>
                    </div>
                    <div className="input-box">
                        <span className='icon'></span>
                        <input type="email" value={email} onChange={e => setEmail(e.target.value)} required />
                        <label>Email</label>
                    </div>
                    <div className="input-box">
                        <span className='icon'></span>
                        <input type="password" value={password} onChange={e => setPassword(e.target.value)} required />
                        <label>Password</label>
                    </div>
                    {err && <p className="error-message">{err}</p>}
                    <button type='submit'>Register</button>
                </form>
            </div>
        </section>
    )
}
