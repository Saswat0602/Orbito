import { io, Socket } from 'socket.io-client';

let socket: Socket | null = null;

export const getSocket = (token?: string): Socket => {
  if (!socket) {
    const socketUrl = process.env.NEXT_PUBLIC_SOCKET_URL || 'http://localhost:3000';
    const authToken = token || (typeof window !== 'undefined' ? localStorage.getItem('orbito_access_token') || '' : '');

    socket = io(`${socketUrl}/events`, {
      auth: {
        token: authToken,
      },
      transports: ['websocket', 'polling'],
      autoConnect: true,
      reconnection: true,
      reconnectionAttempts: 5,
      reconnectionDelay: 1000,
    });

    socket.on('connect', () => {
      console.log('Orbito Socket connected:', socket?.id);
    });

    socket.on('disconnect', (reason) => {
      console.log('Orbito Socket disconnected:', reason);
    });

    socket.on('error', (err) => {
      console.error('Orbito Socket error:', err);
    });
  }

  return socket;
};

export const disconnectSocket = () => {
  if (socket) {
    socket.disconnect();
    socket = null;
  }
};
