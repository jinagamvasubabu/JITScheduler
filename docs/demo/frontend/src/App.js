import React, { useState, useEffect } from 'react';
import io from 'socket.io-client';

const socket = io('http://localhost:3000'); // Replace with your backend address

function App() {
  const [alert, setAlert] = useState('');

  useEffect(() => {
    socket.on('alert', (message) => {
      setAlert(message);
    });
  }, []);

  return (
      <div>
        {alert && (
            <div className="alert">
              <p>{alert}</p>
            </div>
        )}
      </div>
  );
}

export default App;