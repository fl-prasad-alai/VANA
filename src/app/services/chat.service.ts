// src/app/services/chat.service.ts

import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable } from 'rxjs';
import { ChatMessage, Conversation, ChatRequest, ChatResponse, TriageResponse } from '../models/chat.models';

@Injectable({
  providedIn: 'root'
})
export class ChatService {
  private apiUrl = '/api';
  private conversationSubject = new BehaviorSubject<Conversation | null>(null);
  public conversation$ = this.conversationSubject.asObservable();

  private messagesSubject = new BehaviorSubject<ChatMessage[]>([]);
  public messages$ = this.messagesSubject.asObservable();

  private isLoadingSubject = new BehaviorSubject<boolean>(false);
  public isLoading$ = this.isLoadingSubject.asObservable();

  constructor(private http: HttpClient) {}

  // Create a new conversation
  createConversation(): Observable<Conversation> {
    return this.http.post<Conversation>(`${this.apiUrl}/conversations`, {});
  }

  // Get conversation by ID
  getConversation(id: string): Observable<Conversation> {
    return this.http.get<Conversation>(`${this.apiUrl}/conversations/${id}`);
  }

  // Send a message and get AI response
  sendMessage(request: ChatRequest): Observable<ChatResponse> {
    this.isLoadingSubject.next(true);
    return this.http.post<ChatResponse>(`${this.apiUrl}/chat`, request);
  }

  // Get conversation messages
  getMessages(conversationId: string): Observable<ChatMessage[]> {
    return this.http.get<ChatMessage[]>(`${this.apiUrl}/conversations/${conversationId}/messages`);
  }

  // Submit GAD-7 triage
  submitTriage(conversationId: string, responses: number[]): Observable<TriageResponse> {
    return this.http.post<TriageResponse>(`${this.apiUrl}/triage`, {
      conversation_id: conversationId,
      responses
    });
  }

  // Update local conversation
  setCurrentConversation(conversation: Conversation): void {
    this.conversationSubject.next(conversation);
  }

  // Update local messages
  setMessages(messages: ChatMessage[]): void {
    this.messagesSubject.next(messages);
  }

  // Add a message to local cache
  addMessage(message: ChatMessage): void {
    const current = this.messagesSubject.value;
    this.messagesSubject.next([...current, message]);
  }

  // Get current conversation
  getCurrentConversation(): Conversation | null {
    return this.conversationSubject.value;
  }

  // Set loading state
  setLoading(loading: boolean): void {
    this.isLoadingSubject.next(loading);
  }
}
