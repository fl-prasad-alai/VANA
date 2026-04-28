// src/app/models/chat.models.ts

export interface ChatMessage {
  id: string;
  conversationId: string;
  sender: 'user' | 'ai';
  content: string;
  sentiment?: number; // -1 (negative) to 1 (positive)
  timestamp: string;
  isEncrypted?: boolean;
}

export interface Conversation {
  id: string;
  userId: string;
  gad7Score?: number;
  status: 'active' | 'archived' | 'closed';
  createdAt: string;
  updatedAt: string;
  messages?: ChatMessage[];
}

export interface TriageResponse {
  gad7Score: number;
  severity: 'minimal' | 'mild' | 'moderate' | 'severe';
  recommendation: string;
}

export interface ChatRequest {
  message: string;
  conversationId: string;
}

export interface ChatResponse {
  id: string;
  conversationId: string;
  response: string;
  sentiment: number;
  providerId: 'groq' | 'gemini';
  timestamp: string;
}
