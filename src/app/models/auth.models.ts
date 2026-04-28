// src/app/models/auth.models.ts

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  confirmPassword: string;
  fullName: string;
}

export interface AuthResponse {
  success: boolean;
  message: string;
  token?: string;
  user?: User;
}

export interface User {
  id: string;
  email: string;
  fullName: string;
  createdAt: string;
  encryptionKey?: string;
}

export interface JwtToken {
  token: string;
  expiresIn: number;
}
