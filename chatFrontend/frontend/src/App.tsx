import { useState } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Home from './home/Home';
import { Login } from './login/Login';
import { Register } from './register/Register'

function App() {
    const [user, setUser] = useState("");
    const [chats, setChats] = useState<string[]>([]);
    return (
        <Router>
            <Routes>
                <Route path="/" element={<Login onSelectUser={setUser} onSelectChats={setChats} />} />
                <Route path="/home" element={<Home user={user} chats={chats} />} />
                <Route path="/register" element={<Register />} />
            </Routes>
        </Router>
    );
}

export default App;
