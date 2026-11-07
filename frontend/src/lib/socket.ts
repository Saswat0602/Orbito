import { io, Socket } from 'socket.io-client';
import { SOCKET_URL } from './api/endpoint';

let socket: Socket | null = null;

export const getSocket = (): Socket => {
  if (!socket) {
    socket = io(`${SOCKET_URL}/events`, {
      withCredentials: true, // Automatically sends auth and session cookies
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
