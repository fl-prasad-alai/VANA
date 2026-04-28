// src/app/services/auth.service.ts

import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable, tap, catchError } from 'rxjs';
import { LoginRequest, RegisterRequest, AuthResponse, User } from '../models/auth.models';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private apiUrl = '/api/auth';
  private currentUserSubject = new BehaviorSubject<User | null>(this.getUserFromStorage());
  public currentUser$ = this.currentUserSubject.asObservable();

  private isAuthenticatedSubject = new BehaviorSubject<boolean>(this.hasValidToken());
  public isAuthenticated$ = this.isAuthenticatedSubject.asObservable();

  constructor(private http: HttpClient) {}

  login(email: string, password: string): Observable<AuthResponse> {
    const request: LoginRequest = { email, password };
    return this.http.post<AuthResponse>(`${this.apiUrl}/login`, request).pipe(
      tap(response => {
        if (response.success && response.token && response.user) {
          this.setToken(response.token);
          this.setUser(response.user);
          this.isAuthenticatedSubject.next(true);
        }
      }),
      catchError(error => {
        console.error('Login error:', error);
        throw error;
      })
    );
  }

  register(email: string, password: string, fullName: string): Observable<AuthResponse> {
    const request: RegisterRequest = {
      email,
      password,
      confirmPassword: password,
      fullName
    };
    return this.http.post<AuthResponse>(`${this.apiUrl}/register`, request).pipe(
      tap(response => {
        if (response.success && response.token && response.user) {
          this.setToken(response.token);
          this.setUser(response.user);
          this.isAuthenticatedSubject.next(true);
        }
      }),
      catchError(error => {
        console.error('Register error:', error);
        throw error;
      })
    );
  }

  logout(): void {
    localStorage.removeItem('access_token');
    sessionStorage.removeItem('access_token');
    localStorage.removeItem('user');
    sessionStorage.removeItem('user');
    this.currentUserSubject.next(null);
    this.isAuthenticatedSubject.next(false);
  }

  getToken(): string | null {
    return localStorage.getItem('access_token') || sessionStorage.getItem('access_token');
  }

  hasValidToken(): boolean {
    const token = this.getToken();
    if (!token) return false;
    
    try {
      // Simple token existence check; in production, validate JWT expiration
      const payload = JSON.parse(atob(token.split('.')[1]));
      return payload.exp * 1000 > Date.now();
    } catch {
      return false;
    }
  }

  getCurrentUser(): User | null {
    return this.currentUserSubject.value;
  }

  private setToken(token: string): void {
    localStorage.setItem('access_token', token);
    sessionStorage.setItem('access_token', token);
  }

  private setUser(user: User): void {
    localStorage.setItem('user', JSON.stringify(user));
    sessionStorage.setItem('user', JSON.stringify(user));
    this.currentUserSubject.next(user);
  }

  private getUserFromStorage(): User | null {
    try {
      const userJson = localStorage.getItem('user') || sessionStorage.getItem('user');
      return userJson ? JSON.parse(userJson) : null;
    } catch {
      return null;
    }
  }
}
