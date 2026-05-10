import 'regenerator-runtime/runtime';

import React, { useState, useEffect, useRef, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Leaf, Send, Mic, MicOff, Plus, LogOut, Settings, BarChart3, ChevronRight,
  Heart, AlertTriangle, MessageCircle, Sparkles, Volume2, VolumeX
} from 'lucide-react';
import { Howl } from 'howler';
import { useAuth } from '../hooks/useAuth';
import { useTheme } from '../contexts/ThemeContext';
import CinematicBackground from '../components/CinematicBackground';
import ThemeToggle from '../components/ThemeToggle';
import VitalityCore from '../components/VitalityCore';
import ReactMarkdown from 'react-markdown';
import SpeechRecognition, { useSpeechRecognition } from 'react-speech-recognition';

const SOUNDSCAPE_LIBRARY: Record<string, string> = {
  forest_morning: 'https://raw.githubusercontent.com/CS42org/Nature-sounds/main/Sounds/01_bird/1.mp3',
  monsoon_rain: 'https://raw.githubusercontent.com/CS42org/Nature-sounds/main/Sounds/04_rain/4.mp3',
  mossy_stream: 'https://raw.githubusercontent.com/CS42org/Nature-sounds/main/Sounds/05_stream/13.mp3',
  dusk_valley: 'https://raw.githubusercontent.com/CS42org/Nature-sounds/main/Sounds/01_bird/1.mp3',
  sunlight_canopy: 'https://raw.githubusercontent.com/CS42org/Nature-sounds/main/Sounds/04_rain/4.mp3',
};

interface Message {
  id: string;
  sender: 'user' | 'bot';
  text: string;
  timestamp: Date;
  sentiment?: 'positive' | 'neutral' | 'negative' | 'critical';
}

interface Session {
  id: string;
  label: string;
  date: string;
  active?: boolean;
}

const INITIAL_SESSIONS: Session[] = [
  { id: '1', label: "Today's Session", date: 'Today', active: true },
  { id: '2', label: 'Anxiety & Work Stress', date: 'Yesterday' },
  { id: '3', label: 'Sleep Patterns', date: '3 days ago' },
  { id: '4', label: 'Gratitude Practice', date: 'Last week' },
];

const FADE_UP = {
  hidden: { opacity: 0, y: 14 },
  show: { opacity: 1, y: 0, transition: { type: 'spring' as const, stiffness: 300, damping: 26 } },
};

const SentimentDot = React.memo(({ sentiment }: { sentiment?: string }) => {
  if (!sentiment || sentiment === 'neutral') return <MessageCircle className="w-4 h-4 opacity-60" />;
  if (sentiment === 'positive') return <Heart className="w-4 h-4 text-emerald-400" />;
  if (sentiment === 'critical') return <AlertTriangle className="w-4 h-4 text-red-400 animate-pulse" />;
  return <Sparkles className="w-4 h-4 text-violet-400" />;
});

const TypingDots = () => (
  <div className="flex items-center gap-1 px-1 py-0.5">
    {[0, 1, 2].map((i) => (
      <motion.span
        key={i}
        animate={{ y: [0, -6, 0], opacity: [0.4, 1, 0.4] }}
        transition={{ duration: 0.8, repeat: Infinity, delay: i * 0.15 }}
        className="w-2 h-2 rounded-full bg-current inline-block"
      />
    ))}
  </div>
);

const MemoizedMarkdown = React.memo(({ content }: { content: string }) => (
  <div className="text-sm leading-relaxed space-y-3 break-words overflow-hidden">
    <ReactMarkdown
      components={{
        h2: ({ ...props }) => <h2 className="text-lg font-bold mt-2 mb-1" {...props} />,
        ul: ({ ...props }) => <ul className="list-disc pl-5 space-y-2 my-2" {...props} />,
        li: ({ ...props }) => <li className="leading-relaxed" {...props} />,
        p: ({ ...props }) => <p className="leading-relaxed" {...props} />,
        hr: ({ ...props }) => <hr className="my-4 border-current opacity-20" {...props} />,
        strong: ({ ...props }) => <strong className="font-bold text-current" {...props} />,
      }}
    >
      {content}
    </ReactMarkdown>
  </div>
));

const VoiceInput = ({ onSend, loading, accentBg }: { onSend: (text: string, isVoice: boolean) => void, loading: boolean, accentBg: string }) => {
  const { transcript, listening, resetTranscript, browserSupportsSpeechRecognition } = useSpeechRecognition();
  const silenceTimerRef = useRef<any>(null);

  const handleStop = useCallback(() => {
    console.log('VANA-VOICE: Executing manual abort...');
    SpeechRecognition.abortListening();
    
    if (transcript.trim()) {
      onSend(transcript, true);
    }
    
    resetTranscript();
    if (silenceTimerRef.current) clearTimeout(silenceTimerRef.current);
  }, [transcript, onSend, resetTranscript]);

  const startListening = () => {
    resetTranscript();
    console.log('VANA-VOICE: Attempting to start listening...');
    
    try {
      SpeechRecognition.startListening({ 
        continuous: true, 
        language: 'en-IN' 
      }).catch(err => {
        console.warn('VANA-VOICE: en-IN failed, falling back to default', err);
        SpeechRecognition.startListening({ continuous: false });
      });
    } catch (e) {
      console.error('VANA-VOICE: Critical failure in startListening', e);
    }
  };

  const toggleListening = () => {
    if (listening) {
      console.log('VANA-VOICE: Stopping...');
      handleStop();
    } else {
      startListening();
    }
  };

  useEffect(() => {
    if (listening) {
      if (silenceTimerRef.current) clearTimeout(silenceTimerRef.current);
      silenceTimerRef.current = setTimeout(() => {
        handleStop();
      }, 10000);
    }
    return () => {
      if (silenceTimerRef.current) clearTimeout(silenceTimerRef.current);
    };
  }, [transcript, listening, handleStop]);

  // Status reporter for production debugging
  const getStatusText = () => {
    if (listening) return 'Listening...';
    if (!browserSupportsSpeechRecognition) return 'Not Supported';
    return 'Ready';
  };

  if (!browserSupportsSpeechRecognition) {
    console.warn('VANA-VOICE: Browser does not support speech recognition');
    return (
      <button 
        disabled 
        title="Speech recognition not supported in this browser"
        className="flex-shrink-0 w-11 h-11 rounded-2xl flex items-center justify-center bg-black/5 text-zinc-400 opacity-50 cursor-not-allowed"
      >
        <MicOff className="w-5 h-5" />
      </button>
    );
  }

  // Production Check: Ensure we are in a secure context (HTTPS)
  if (!window.isSecureContext) {
    console.error('VANA-VOICE: Microphone access requires HTTPS on production domains.');
  }

  return (
    <motion.button
      whileHover={{ scale: 1.05 }}
      whileTap={{ scale: 0.95 }}
      onClick={(e) => {
        e.preventDefault();
        console.log('VANA-VOICE: Toggle interaction detected');
        toggleListening();
      }}
      disabled={loading}
      className={`
        flex-shrink-0 px-4 h-11 rounded-2xl flex items-center justify-center transition-all font-bold text-xs tracking-wider uppercase
        ${listening
          ? `${accentBg} text-white shadow-lg animate-pulse`
          : 'dark:bg-white/5 bg-black/5 text-zinc-400 hover:text-zinc-200'}
      `}
      title={listening ? "Click to stop" : "Click to speak"}
    >
      <div className="relative">
        {listening ? (
          <span className="flex items-center gap-2">
            <Mic className="w-4 h-4 animate-bounce text-white"/> 
            <span className="hidden md:inline text-[10px] font-bold tracking-tight">🌳  Listening...</span>
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Mic className="w-4 h-4"/> 
            <span className="hidden md:inline text-[10px] tracking-tight">Speak to VANA</span>
          </span>
        )}
      </div>
    </motion.button>
  );
};

export const DashboardPage: React.FC = () => {
  const navigate = useNavigate();
  const { user, logout, isAuthenticated } = useAuth();
  const { theme } = useTheme();
  const isGreening = theme === 'greening';
  const isDark = theme === 'dark';

  const [messages, setMessages] = useState<Message[]>([
    {
      id: '1',
      sender: 'bot',
      text: `Hello${user?.fullName ? `, ${user.fullName.split(' ')[0]}` : ''}. I'm VANA — your private, mindful AI companion. How are you feeling today?`,
      timestamp: new Date(),
      sentiment: 'neutral',
    },
  ]);

  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [sessions, setSessions] = useState<Session[]>(INITIAL_SESSIONS);
  const [activeSession, setActiveSession] = useState('1');
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [scrolled, setScrolled] = useState(false);
  const [isReflecting, setIsReflecting] = useState(false);
  
  // Audio State
  const [musicEnabled, setMusicEnabled] = useState(() => localStorage.getItem('vana_music_enabled') !== 'false');
  const [currentTrack, setCurrentTrack] = useState('forest_morning');
  const soundRef = useRef<Howl | null>(null);
  const { listening } = useSpeechRecognition();

  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const chatContainerRef = useRef<HTMLDivElement>(null);
  
  // Initial Modal State: Prevents audio issues and sets the vibe
  const [showInitialModal, setShowInitialModal] = useState(true);

  useEffect(() => {
    if (!isAuthenticated) navigate('/login');
  }, [isAuthenticated, navigate]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, loading]);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 10);
    window.addEventListener('scroll', onScroll);
    return () => window.removeEventListener('scroll', onScroll);
  }, []);

  useEffect(() => {
    if (!isAuthenticated) navigate('/login');
  }, [isAuthenticated, navigate]);

  // Audio Engine Management
  useEffect(() => {
    if (!musicEnabled || showInitialModal) {
      if (soundRef.current) {
        soundRef.current.stop();
      }
      return;
    }

    if (!soundRef.current) {
      soundRef.current = new Howl({
        src: [SOUNDSCAPE_LIBRARY[currentTrack]],
        html5: false,
        loop: true,
        volume: 0.3,
      });
      soundRef.current.play();
    } else {
      const oldSound = soundRef.current;
      const newSound = new Howl({
        src: [SOUNDSCAPE_LIBRARY[currentTrack]],
        html5: false,
        loop: true,
        volume: 0,
      });
      
      newSound.play();
      newSound.fade(0, 0.3, 3000);
      oldSound.fade(oldSound.volume(), 0, 3000);
      setTimeout(() => {
        oldSound.stop();
        oldSound.unload();
      }, 3500);
      soundRef.current = newSound;
    }

    return () => {
      if (soundRef.current) {
        soundRef.current.stop();
        soundRef.current.unload();
      }
    };
  }, [musicEnabled, currentTrack, showInitialModal]);

  const selectInitialSoundscape = (track: string) => {
    setCurrentTrack(track);
    setShowInitialModal(false);
    // User gesture occurs here, allowing immediate play
    if (typeof Howler !== 'undefined' && Howler.ctx) {
      Howler.ctx.resume();
    }
  };

  // Voice Ducking Logic
  useEffect(() => {
    if (soundRef.current && musicEnabled) {
      if (listening) {
        soundRef.current.fade(soundRef.current.volume(), 0, 500);
      } else {
        soundRef.current.fade(soundRef.current.volume(), 0.2, 1500);
      }
    }
  }, [listening, musicEnabled]);

  const toggleMusic = () => {
    const newState = !musicEnabled;
    setMusicEnabled(newState);
    localStorage.setItem('vana_music_enabled', String(newState));
    
    if (newState && soundRef.current && !soundRef.current.playing()) {
      soundRef.current.play();
    }
  };

  const sendMessage = async (text: string = input, isVoiceInput: boolean = false) => {
    if (!text.trim() || loading) return;
    const userMsg: Message = { id: Date.now().toString(), sender: 'user', text, timestamp: new Date() };
    setMessages(prev => [...prev, userMsg]);
    if (!isVoiceInput) {
      setInput('');
    }
    if (isVoiceInput) setIsReflecting(true);
    setLoading(true);
    inputRef.current?.focus();

    try {
      const response = await axios.post('/chat', {
        message: text,
        conversationId: activeSession,
        userId: user?.id,
        isVoiceInput: isVoiceInput,
      });

      let sentiment: 'neutral' | 'positive' | 'negative' | 'critical' = 'neutral';
      if (response.data.crisis) {
        sentiment = 'critical';
      } else if (response.data.sentiment_score > 0.6) {
        sentiment = 'positive';
      } else if (response.data.sentiment_score < 0.4) {
        sentiment = 'negative';
      }

      const botMsg: Message = { 
        id: (Date.now() + 1).toString(), 
        sender: 'bot', 
        text: response.data.text, 
        timestamp: new Date(), 
        sentiment: sentiment 
      };
      
      setMessages(prev => [...prev, botMsg]);
    } catch (error: any) {
      console.error('Chat error:', error);
      const isRateLimit = error.response?.status === 429;
      const errorText = isRateLimit 
        ? "I'm receiving too many messages right now. Please take a deep breath and wait a moment."
        : "I'm having trouble connecting right now. Please try again in a moment.";

      const botMsg: Message = { id: (Date.now() + 1).toString(), sender: 'bot', text: errorText, timestamp: new Date(), sentiment: 'critical' };
      setMessages(prev => [...prev, botMsg]);
    } finally {
      setLoading(false);
      setIsReflecting(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); sendMessage(); }
  };

  const startNewChat = () => {
    const id = Date.now().toString();
    setSessions(prev => [{ id, label: 'New Conversation', date: 'Just now', active: true }, ...prev.map(s => ({ ...s, active: false }))]);
    setActiveSession(id);
    setMessages([{ id: '1', sender: 'bot', text: "Hi again! What's on your mind today?", timestamp: new Date(), sentiment: 'neutral' }]);
    setSidebarOpen(false);
  };

  const accentClass = isGreening ? 'text-emerald-400' : isDark ? 'text-violet-400' : 'text-[#e11d48]';
  const accentBg = isGreening ? 'bg-emerald-600' : isDark ? 'bg-violet-600' : 'bg-[#e11d48]';

  return (
    <div className={`relative flex flex-col h-screen overflow-hidden font-sans ${isGreening ? 'bg-[#022c22] text-[#ecfdf5]' : 'bg-slate-50 dark:bg-black text-zinc-900 dark:text-zinc-100'}`}>
      <CinematicBackground />

      <AnimatePresence>
        {showInitialModal && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-[100] flex items-end md:items-center justify-center bg-black/50 backdrop-blur-md px-0 md:px-4"
          >
            <motion.div
              initial={{ y: "100%" }}
              animate={{ y: 0 }}
              exit={{ y: "100%" }}
              transition={{ type: "spring", damping: 25, stiffness: 200 }}
              className="w-full md:max-w-2xl bg-white/10 dark:bg-zinc-900/40 border-t md:border border-white/20 rounded-t-[2.5rem] md:rounded-[2rem] p-10 md:p-12 shadow-2xl text-center relative"
            >
              {/* Mobile Handle */}
              <div className="w-12 h-1.5 bg-white/20 rounded-full mx-auto mb-8 md:hidden" />

              <h2 className="text-2xl md:text-4xl font-bold text-white mb-2 md:mb-4 tracking-tight">
                Welcome to VANA
              </h2>
              <p className="text-zinc-400 text-sm md:text-lg mb-10 md:mb-12 max-w-md mx-auto leading-relaxed">
                Choose your soundscape
              </p>

              <div className="grid grid-cols-3 md:grid-cols-3 gap-3 md:gap-4">
                {[
                  { id: 'forest_morning', label: 'Deep Forest', icon: Leaf, desc: 'Birds & Soft Wind' },
                  { id: 'monsoon_rain', label: 'Rain Sanctuary', icon: MessageCircle, desc: 'Gentle Monsoon' },
                  { id: 'mossy_stream', label: 'Mossy Stream', icon: Heart, desc: 'Flowing Water' }
                ].map((item) => (
                  <button
                    key={item.id}
                    onClick={() => selectInitialSoundscape(item.id)}
                    className="group relative flex flex-col items-center p-3 md:p-6 rounded-2xl md:rounded-3xl bg-white/5 border border-white/10 hover:bg-white/10 hover:border-white/20 transition-all duration-300"
                  >
                    <item.icon className="w-5 h-5 md:w-6 md:h-6 mb-2 md:mb-3 text-emerald-400 group-hover:scale-110 transition-transform" />
                    <span className="text-white text-xs md:text-base font-medium mb-1 md:mb-1">{item.label.split(' ')[0]}</span>
                    <span className="text-zinc-300 text-[10px] md:text-xs group-hover:text-white transition-colors text-center leading-tight">{item.desc}</span>
                  </button>
                ))}
              </div>
              
              <p className="hidden md:block mt-10 text-zinc-500 text-xs italic">
                Click a path to enter the stillness.
              </p>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* NAVBAR */}
      <motion.header
        className={`
          fixed top-0 inset-x-0 z-40 flex items-center justify-between px-4 sm:px-8 h-16
          transition-all duration-300
          ${scrolled
            ? isGreening
              ? 'bg-[#022c22]/90 backdrop-blur-xl border-b border-emerald-500/10 shadow-lg'
              : 'bg-white/80 dark:bg-black/80 backdrop-blur-xl border-b border-black/5 dark:border-white/5 shadow-lg'
            : 'bg-transparent'}
        `}
      >
        <div className="flex items-center gap-3">
          <button className="lg:hidden p-2 rounded-xl dark:bg-white/5 bg-black/5 mr-1" onClick={() => setSidebarOpen(v => !v)}>
            <ChevronRight className={`w-4 h-4 transition-transform ${sidebarOpen ? 'rotate-180' : ''}`} />
          </button>
          <div className={`w-8 h-8 rounded-xl flex items-center justify-center shadow ${isGreening ? 'bg-emerald-600' : 'bg-[#e11d48]'}`}>
            <Leaf className="w-4 h-4 text-white" />
          </div>
          <span className="font-black tracking-tighter font-heading text-lg">VANA <span className={accentClass}>MIND</span></span>
        </div>
        <div className="flex items-center gap-3">
          <span className="hidden sm:block text-xs font-mono text-zinc-500 dark:text-zinc-600 tracking-wider">
            {user?.fullName?.split(' ')[0] || 'Guest'}
          </span>
          <div className="hidden sm:block w-px h-5 bg-white/10" />
          <motion.button
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            onClick={toggleMusic}
            className="p-2 rounded-xl dark:bg-white/5 bg-black/5 text-zinc-500 hover:text-zinc-300 transition-colors"
            title={musicEnabled ? "Mute Soundscape" : "Unmute Soundscape"}
          >
            {musicEnabled ? <Volume2 className="w-4 h-4" /> : <VolumeX className="w-4 h-4" />}
          </motion.button>
          <ThemeToggle />
          <button onClick={logout} className="ml-2 flex items-center gap-1.5 text-xs font-semibold text-zinc-500 dark:text-zinc-400 hover:text-red-400 transition-colors px-3 py-2 rounded-xl hover:bg-red-500/10">
            <LogOut className="w-4 h-4" />
            <span className="hidden sm:block">Sign out</span>
          </button>
        </div>
      </motion.header>

      <div className="flex flex-1 overflow-hidden pt-16">
        {/* SIDEBAR */}
        <AnimatePresence>
          {(sidebarOpen || true) && (
            <motion.aside
              initial={{ x: -280, opacity: 0 }} animate={{ x: 0, opacity: 1 }} exit={{ x: -280, opacity: 0 }}
              transition={{ type: 'spring', stiffness: 300, damping: 30 }}
              className={`${sidebarOpen ? 'fixed inset-y-0 left-0 pt-16 z-30' : 'hidden lg:flex lg:flex-col'} w-72 flex-shrink-0 overflow-y-auto ${isGreening ? 'bg-emerald-950/30 border-r border-emerald-500/10' : 'bg-white/40 dark:bg-white/[.02] border-r border-black/5 dark:border-white/5'} backdrop-blur-2xl`}
            >
              <div className="flex flex-col items-center py-8 border-b border-white/5">
                <VitalityCore />
                <p className={`mt-3 text-xs font-mono tracking-widest uppercase ${accentClass}`}>Vitality Core</p>
              </div>
              <div className="p-4">
                <motion.button whileHover={{ scale: 1.02 }} whileTap={{ scale: 0.98 }} onClick={startNewChat} className={`w-full flex items-center justify-center gap-2 py-3 rounded-2xl text-white text-xs font-bold uppercase tracking-widest shadow-xl ${accentBg}`}>
                  <Plus className="w-4 h-4" /> New Conversation
                </motion.button>
              </div>
              <div className="px-4 pb-4 flex-1 space-y-1 overflow-y-auto">
                <p className="text-[10px] font-bold tracking-[0.2em] text-zinc-500 uppercase mb-3 ml-2">Recent Sessions</p>
                {sessions.map((s) => (
                  <motion.button key={s.id} whileHover={{ x: 4 }} onClick={() => { setActiveSession(s.id); setSidebarOpen(false); }} className={`w-full text-left flex items-center gap-3 p-3 rounded-2xl transition-all text-sm ${activeSession === s.id ? isGreening ? 'bg-emerald-500/15 border border-emerald-500/20 text-emerald-100' : 'bg-white/[.06] border border-white/10 dark:text-zinc-100 text-zinc-900' : 'text-zinc-500 dark:text-zinc-500 hover:bg-white/5 hover:text-zinc-300'}`}>
                    <MessageCircle className="w-4 h-4 flex-shrink-0 opacity-60" />
                    <div className="min-w-0">
                      <p className="font-medium truncate">{s.label}</p>
                      <p className="text-[10px] text-zinc-500 font-mono">{s.date}</p>
                    </div>
                    {activeSession === s.id && <ChevronRight className={`w-3 h-3 ml-auto ${accentClass}`} />}
                  </motion.button>
                ))}
              </div>
              <div className="p-4 border-t border-white/5 space-y-2">
                <button className="w-full flex items-center gap-2 p-2.5 rounded-xl text-xs text-zinc-500 hover:text-zinc-300 hover:bg-white/5 transition-all">
                  <BarChart3 className="w-4 h-4" /> Session Reports
                </button>
                <button className="w-full flex items-center gap-2 p-2.5 rounded-xl text-xs text-zinc-500 hover:text-zinc-300 hover:bg-white/5 transition-all">
                  <Settings className="w-4 h-4" /> Settings
                </button>
              </div>
            </motion.aside>
          )}
        </AnimatePresence>

        {/* CHAT AREA */}
        <div className="flex-1 flex flex-col overflow-hidden">
          <div ref={chatContainerRef} className="flex-1 overflow-y-auto px-4 sm:px-8 py-6 space-y-6">
            <AnimatePresence initial={false}>
              {messages.map((msg) => (
                <motion.div key={msg.id} variants={FADE_UP} initial="hidden" animate="show" className={`flex gap-3 ${msg.sender === 'user' ? 'flex-row-reverse' : 'flex-row'} max-w-4xl ${msg.sender === 'user' ? 'ml-auto' : 'mr-auto'}`}>
                  <div className={`w-9 h-9 rounded-2xl flex-shrink-0 flex items-center justify-center shadow-lg ${msg.sender === 'bot' ? accentBg : 'bg-zinc-700 dark:bg-zinc-800'}`}>
                    {msg.sender === 'bot' ? <Leaf className="w-4 h-4 text-white" /> : <span className="text-sm font-bold text-white">{user?.fullName?.[0] || 'U'}</span>}
                  </div>
                  <div className="max-w-[85%]">
                    <div className={`relative px-5 py-4 rounded-3xl shadow-lg ${msg.sender === 'user' ? `${isGreening ? 'bg-emerald-600' : 'bg-[#e11d48]'} text-white rounded-tr-lg` : `${isGreening ? 'bg-emerald-950/40' : 'bg-white/40 dark:bg-white/5'} backdrop-blur-xl border ${isGreening ? 'border-emerald-500/15' : 'border-black/5 dark:border-white/10'} text-zinc-900 dark:text-zinc-100 rounded-tl-lg`}`}>
                      {msg.sender === 'bot' && <span className="absolute -top-1 -left-1"><SentimentDot sentiment={msg.sentiment} /></span>}
                      <MemoizedMarkdown content={msg.text} />
                    </div>
                    <p className={`text-[10px] mt-1.5 font-mono opacity-40 ${msg.sender === 'user' ? 'text-right' : 'text-left'}`}>
                      {msg.timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                    </p>
                  </div>
                </motion.div>
              ))}
              {loading && (
                <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} className="flex gap-3 mr-auto max-w-4xl">
                  <div className={`w-9 h-9 rounded-2xl flex-shrink-0 flex items-center justify-center shadow-lg ${accentBg}`}>
                    <Leaf className="w-4 h-4 text-white" />
                  </div>
                  <div className={`px-5 py-4 rounded-3xl shadow-lg ${isGreening ? 'bg-emerald-950/40' : 'bg-white/40 dark:bg-white/5'} backdrop-blur-xl border ${isGreening ? 'border-emerald-500/15' : 'border-black/5 dark:border-white/10'} text-zinc-900 dark:text-zinc-100 rounded-tl-lg`}>
                    {isReflecting && <p className={`text-[10px] mb-2 font-mono tracking-widest uppercase opacity-60 ${accentClass}`}>VANA is reflecting...</p>}
                    <TypingDots />
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
            <div ref={messagesEndRef} />
          </div>

          {/* Input area */}
          <div className={`px-4 sm:px-8 pb-6 pt-4 border-t ${isGreening ? 'border-emerald-500/10' : 'border-black/5 dark:border-white/5'}`}>
            <AnimatePresence>
              {listening && musicEnabled && (
                <motion.p
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: 10 }}
                  className="text-[10px] font-mono text-center mb-2 text-zinc-500 italic"
                >
                  VANA is listening in the stillness...
                </motion.p>
              )}
            </AnimatePresence>
            <p className={`text-[10px] font-mono text-center mb-4 opacity-50 ${accentClass}`}>
              💚 VANA supports, not replaces, professional care. For immediate help, call 988 (US) or iCall: 9152987821 (India).
            </p>
            <div className={`flex items-end gap-3 p-3 rounded-3xl shadow-xl transition-all duration-300 ${isGreening ? 'bg-emerald-950/30 backdrop-blur-xl border border-emerald-500/15' : 'bg-white/60 dark:bg-white/5 backdrop-blur-xl border border-black/5 dark:border-white/10'}`}>
              
              <VoiceInput onSend={sendMessage} loading={loading} accentBg={accentBg} />

              <input ref={inputRef} type="text" value={input} onChange={e => setInput(e.target.value)} onKeyDown={handleKeyDown} placeholder={'Share what\'s on your mind...'} disabled={loading} className={`flex-1 bg-transparent text-sm placeholder-zinc-400 dark:placeholder-zinc-600 focus:outline-none resize-none py-2 ml-2 ${isGreening ? 'text-emerald-50' : 'text-zinc-900 dark:text-zinc-100'}`} />
              <motion.button whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.95 }} onClick={() => sendMessage(input, false)} disabled={loading || !input.trim()} className={`flex-shrink-0 w-11 h-11 rounded-2xl flex items-center justify-center transition-all shadow-lg ${input.trim() && !loading ? `${accentBg} text-white` : 'dark:bg-white/5 bg-black/5 text-zinc-400 cursor-not-allowed opacity-50'}`}>
                <Send className="w-5 h-5" />
              </motion.button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
