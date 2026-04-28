// src/app/services/voice.service.ts

import { Injectable, NgZone } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';

declare global {
  interface Window {
    webkitSpeechRecognition: any;
    SpeechRecognition: any;
    SpeechSynthesisUtterance: any;
  }
}

@Injectable({
  providedIn: 'root'
})
export class VoiceService {
  private recognition: any;
  private synthesis = window.speechSynthesis;

  private isListeningSubject = new BehaviorSubject<boolean>(false);
  public isListening$ = this.isListeningSubject.asObservable();

  private transcriptSubject = new BehaviorSubject<string>('');
  public transcript$ = this.transcriptSubject.asObservable();

  private interimTranscriptSubject = new BehaviorSubject<string>('');
  public interimTranscript$ = this.interimTranscriptSubject.asObservable();

  private isSpeakingSubject = new BehaviorSubject<boolean>(false);
  public isSpeaking$ = this.isSpeakingSubject.asObservable();

  private errorSubject = new BehaviorSubject<string | null>(null);
  public error$ = this.errorSubject.asObservable();

  constructor(private ngZone: NgZone) {
    this.initializeRecognition();
  }

  private initializeRecognition(): void {
    const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
    if (!SpeechRecognition) {
      this.errorSubject.next('Speech Recognition API is not supported in this browser.');
      return;
    }

    this.recognition = new SpeechRecognition();
    this.recognition.continuous = false;
    this.recognition.interimResults = true;
    this.recognition.language = 'en-US';

    // Speech recognition event handlers
    this.recognition.onstart = () => {
      this.ngZone.run(() => {
        this.isListeningSubject.next(true);
        this.errorSubject.next(null);
      });
    };

    this.recognition.onend = () => {
      this.ngZone.run(() => {
        this.isListeningSubject.next(false);
      });
    };

    this.recognition.onresult = (event: any) => {
      this.ngZone.run(() => {
        let interim = '';
        let final = '';

        for (let i = event.resultIndex; i < event.results.length; i++) {
          const transcript = event.results[i][0].transcript;

          if (event.results[i].isFinal) {
            final += transcript + ' ';
          } else {
            interim += transcript;
          }
        }

        if (interim) {
          this.interimTranscriptSubject.next(interim);
        }

        if (final) {
          this.transcriptSubject.next(final.trim());
          this.interimTranscriptSubject.next('');
        }
      });
    };

    this.recognition.onerror = (event: any) => {
      this.ngZone.run(() => {
        this.errorSubject.next(`Speech recognition error: ${event.error}`);
      });
    };
  }

  // Start listening
  startListening(): void {
    if (!this.recognition) {
      this.errorSubject.next('Speech Recognition is not available.');
      return;
    }

    this.transcriptSubject.next('');
    this.interimTranscriptSubject.next('');
    this.recognition.start();
  }

  // Stop listening
  stopListening(): void {
    if (this.recognition) {
      this.recognition.stop();
    }
  }

  // Get the final transcript
  getTranscript(): string {
    return this.transcriptSubject.value;
  }

  // Clear transcript
  clearTranscript(): void {
    this.transcriptSubject.next('');
  }

  // Text-to-Speech
  speak(text: string, rate: number = 1, pitch: number = 1): void {
    if (!this.synthesis) {
      this.errorSubject.next('Speech Synthesis is not available.');
      return;
    }

    // Cancel any ongoing speech
    this.synthesis.cancel();

    const utterance = new window.SpeechSynthesisUtterance(text);
    utterance.rate = rate;
    utterance.pitch = pitch;
    utterance.lang = 'en-US';

    utterance.onstart = () => {
      this.ngZone.run(() => {
        this.isSpeakingSubject.next(true);
      });
    };

    utterance.onend = () => {
      this.ngZone.run(() => {
        this.isSpeakingSubject.next(false);
      });
    };

    utterance.onerror = (event: any) => {
      this.ngZone.run(() => {
        this.errorSubject.next(`Speech synthesis error: ${event.error}`);
        this.isSpeakingSubject.next(false);
      });
    };

    this.synthesis.speak(utterance);
  }

  // Stop speaking
  stopSpeaking(): void {
    if (this.synthesis) {
      this.synthesis.cancel();
      this.isSpeakingSubject.next(false);
    }
  }

  // Check if speech recognition is supported
  isSupported(): boolean {
    return !!(window.SpeechRecognition || window.webkitSpeechRecognition);
  }

  // Check if speech synthesis is supported
  isSynthesisSupported(): boolean {
    return !!window.speechSynthesis;
  }

  // Get available voices
  getVoices(): SpeechSynthesisVoice[] {
    return this.synthesis.getVoices();
  }

  // Set voice for synthesis
  setVoice(voiceName: string): void {
    const voices = this.getVoices();
    const selectedVoice = voices.find(v => v.name === voiceName);
    if (selectedVoice) {
      // Store preference or use when speaking
    }
  }
}
