import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Leaf, Send, Mic, MicOff, Plus, LogOut, Settings, BarChart3, ChevronRight, X,
  Heart, AlertTriangle, MessageCircle, Sparkles,
} from 'lucide-react';
import { useAuth } from '../hooks/useAuth';
import { useTheme } from '../contexts/ThemeContext';
import CinematicBackground from '../components/CinematicBackground';
import ThemeToggle from '../components/ThemeToggle';
import VitalityCore from '../components/VitalityCore';

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

function SentimentDot({ sentiment }: { sentiment?: string }) {
  if (!sentiment || sentiment === 'neutral') return <MessageCircle className="w-4 h-4 opacity-60" />;
  if (sentiment === 'positive') return <Heart className="w-4 h-4 text-emerald-400" />;
  if (sentiment === 'critical') return <AlertTriangle className="w-4 h-4 text-red-400 animate-pulse" />;
  return <Sparkles className="w-4 h-4 text-violet-400" />;
}

function TypingDots() {
  return (
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
}

export const DashboardPage: React.FC = () => {
  const navigate = useNavigate();
  const { user, logout, isAuthenticated } = useAuth();
  const { theme, setTheme } = useTheme();

  const isGreening = theme === 'greening';
  const isDark = theme === 'dark';

  const [messages, setMessages] = useState<Message[]>([
    {
      id: '1',
      sender: 'bot',
      text: `Hello${user?.fullName ? `, ${user.fullName.split(' ')[0]}` : ''}. I'm VANA — your private, mindful AI companion. How are you feeling today? Feel free to share what's on your mind.`,
      timestamp: new Date(),
      sentiment: 'neutral',
    },
  ]);

  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [isListening, setIsListening] = useState(false);
  const [sessions, setSessions] = useState<Session[]>(INITIAL_SESSIONS);
  const [activeSession, setActiveSession] = useState('1');
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [scrolled, setScrolled] = useState(false);

  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const chatContainerRef = useRef<HTMLDivElement>(null);

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

  const sendMessage = async (text: string = input) => {
    if (!text.trim() || loading) return;
    const userMsg: Message = { id: Date.now().toString(), sender: 'user', text, timestamp: new Date() };
    setMessages(prev => [...prev, userMsg]);
    setInput('');
    setLoading(true);
    inputRef.current?.focus();

    await new Promise(r => setTimeout(r, 1200 + Math.random() * 600));

    const responses = [
      { text: "That sounds really meaningful. Could you tell me more about what triggered that feeling?", sentiment: 'positive' as const },
      { text: "I hear you, and I want you to know that what you're experiencing is valid. Let's explore this together.", sentiment: 'neutral' as const },
      { text: "Thank you for sharing that with me. It takes courage to open up. How long have you been feeling this way?", sentiment: 'neutral' as const },
      { text: "I'm here with you through this. What does your body feel like right now as you think about this?", sentiment: 'positive' as const },
    ];
    const chosen = responses[Math.floor(Math.random() * responses.length)];
    const botMsg: Message = { id: (Date.now() + 1).toString(), sender: 'bot', text: chosen.text, timestamp: new Date(), sentiment: chosen.sentiment };
    setMessages(prev => [...prev, botMsg]);
    setLoading(false);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); sendMessage(); }
  };

  const handleVoice = () => {
    if (!('webkitSpeechRecognition' in window || 'SpeechRecognition' in window)) {
      alert('Speech recognition not supported in this browser.');
      return;
    }
    setIsListening(v => !v);
    if (!isListening) setTimeout(() => setIsListening(false), 4000);
  };

  const handleLogout = () => { logout(); navigate('/login'); };

  const startNewChat = () => {
    const id = Date.now().toString();
    setSessions(prev => [{ id, label: 'New Conversation', date: 'Just now', active: true }, ...prev.map(s => ({ ...s, active: false }))]);
    setActiveSession(id);
    setMessages([{ id: '1', sender: 'bot', text: "Hi again! I'm here whenever you're ready. What's on your mind today?", timestamp: new Date(), sentiment: 'neutral' }]);
    setSidebarOpen(false);
  };

  const accentClass = isGreening ? 'text-emerald-400' : isDark ? 'text-violet-400' : 'text-[#e11d48]';
  const accentBg = isGreening ? 'bg-emerald-600' : isDark ? 'bg-violet-600' : 'bg-[#e11d48]';
  const accentBorder = isGreening ? 'border-emerald-500/20' : isDark ? 'border-violet-500/20' : 'border-rose-500/20';

  return (
    <div className={`relative flex flex-col h-screen overflow-hidden font-sans ${isGreening ? 'bg-[#022c22] text-[#ecfdf5]' : 'bg-slate-50 dark:bg-black text-zinc-900 dark:text-zinc-100'}`}>
      <CinematicBackground />

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
        {/* Left */}
        <div className="flex items-center gap-3">
          <button
            className="lg:hidden p-2 rounded-xl dark:bg-white/5 bg-black/5 mr-1"
            onClick={() => setSidebarOpen(v => !v)}
          >
            <ChevronRight className={`w-4 h-4 transition-transform ${sidebarOpen ? 'rotate-180' : ''}`} />
          </button>
          <div className={`w-8 h-8 rounded-xl flex items-center justify-center shadow ${isGreening ? 'bg-emerald-600' : 'bg-[#e11d48]'}`}>
            <Leaf className="w-4 h-4 text-white" />
          </div>
          <span className="font-black tracking-tighter font-heading text-lg">
            VANA <span className={accentClass}>MIND</span>
          </span>
        </div>

        {/* Right */}
        <div className="flex items-center gap-3">
          <span className="hidden sm:block text-xs font-mono text-zinc-500 dark:text-zinc-600 tracking-wider">
            {user?.fullName?.split(' ')[0] || 'Guest'}
          </span>
          <div className="hidden sm:block w-px h-5 bg-white/10" />
          <ThemeToggle />
          <button
            onClick={handleLogout}
            className="ml-2 flex items-center gap-1.5 text-xs font-semibold text-zinc-500 dark:text-zinc-400 hover:text-red-400 transition-colors px-3 py-2 rounded-xl hover:bg-red-500/10"
          >
            <LogOut className="w-4 h-4" />
            <span className="hidden sm:block">Sign out</span>
          </button>
        </div>
      </motion.header>

      {/* MAIN LAYOUT */}
      <div className="flex flex-1 overflow-hidden pt-16">

        {/* SIDEBAR */}
        <AnimatePresence>
          {(sidebarOpen || true) && (
            <motion.aside
              initial={{ x: -280, opacity: 0 }}
              animate={{ x: 0, opacity: 1 }}
              exit={{ x: -280, opacity: 0 }}
              transition={{ type: 'spring', stiffness: 300, damping: 30 }}
              className={`
                ${sidebarOpen ? 'fixed inset-y-0 left-0 pt-16 z-30' : 'hidden lg:flex lg:flex-col'}
                w-72 flex-shrink-0 overflow-y-auto
                ${isGreening
                  ? 'bg-emerald-950/30 border-r border-emerald-500/10'
                  : 'bg-white/40 dark:bg-white/[.02] border-r border-black/5 dark:border-white/5'}
                backdrop-blur-2xl
              `}
            >
              {/* VitalityCore */}
              <div className="flex flex-col items-center py-8 border-b border-white/5">
                <VitalityCore />
                <p className={`mt-3 text-xs font-mono tracking-widest uppercase ${accentClass}`}>Vitality Core</p>
              </div>

              {/* New conversation */}
              <div className="p-4">
                <motion.button
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                  onClick={startNewChat}
                  className={`
                    w-full flex items-center justify-center gap-2 py-3 rounded-2xl
                    text-white text-xs font-bold uppercase tracking-widest shadow-xl
                    ${accentBg}
                  `}
                >
                  <Plus className="w-4 h-4" />
                  New Conversation
                </motion.button>
              </div>

              {/* Session list */}
              <div className="px-4 pb-4 flex-1 space-y-1">
                <p className="text-[10px] font-bold tracking-[0.2em] text-zinc-500 uppercase mb-3 ml-2">Recent Sessions</p>
                {sessions.map((s) => (
                  <motion.button
                    key={s.id}
                    whileHover={{ x: 4 }}
                    onClick={() => { setActiveSession(s.id); setSidebarOpen(false); }}
                    className={`
                      w-full text-left flex items-center gap-3 p-3 rounded-2xl transition-all text-sm
                      ${activeSession === s.id
                        ? isGreening
                          ? 'bg-emerald-500/15 border border-emerald-500/20 text-emerald-100'
                          : 'bg-white/[.06] border border-white/10 dark:text-zinc-100 text-zinc-900'
                        : 'text-zinc-500 dark:text-zinc-500 hover:bg-white/5 hover:text-zinc-300'}
                    `}
                  >
                    <MessageCircle className="w-4 h-4 flex-shrink-0 opacity-60" />
                    <div className="min-w-0">
                      <p className="font-medium truncate">{s.label}</p>
                      <p className="text-[10px] text-zinc-500 font-mono">{s.date}</p>
                    </div>
                    {activeSession === s.id && <ChevronRight className={`w-3 h-3 ml-auto ${accentClass}`} />}
                  </motion.button>
                ))}
              </div>

              {/* Sidebar footer */}
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

        {/* Sidebar overlay for mobile */}
        <AnimatePresence>
          {sidebarOpen && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 z-20 bg-black/60 backdrop-blur-sm lg:hidden"
              onClick={() => setSidebarOpen(false)}
            />
          )}
        </AnimatePresence>

        {/* CHAT AREA */}
        <div className="flex-1 flex flex-col overflow-hidden">

          {/* Messages */}
          <div
            ref={chatContainerRef}
            className="flex-1 overflow-y-auto px-4 sm:px-8 py-6 space-y-4"
          >
            <AnimatePresence initial={false}>
              {messages.map((msg) => (
                <motion.div
                  key={msg.id}
                  variants={FADE_UP}
                  initial="hidden"
                  animate="show"
                  className={`flex gap-3 ${msg.sender === 'user' ? 'flex-row-reverse' : 'flex-row'} max-w-4xl ${msg.sender === 'user' ? 'ml-auto' : 'mr-auto'}`}
                >
                  {/* Avatar */}
                  {msg.sender === 'bot' && (
                    <div className={`w-9 h-9 rounded-2xl flex-shrink-0 flex items-center justify-center shadow-lg ${accentBg}`}>
                      <Leaf className="w-4 h-4 text-white" />
                    </div>
                  )}
                  {msg.sender === 'user' && (
                    <div className="w-9 h-9 rounded-2xl flex-shrink-0 flex items-center justify-center bg-zinc-700 dark:bg-zinc-800 shadow-lg">
                      <span className="text-sm font-bold text-white">{user?.fullName?.[0] || 'U'}</span>
                    </div>
                  )}

                  {/* Bubble */}
                  <div className="max-w-[75%]">
                    <div className={`
                      relative px-5 py-4 rounded-3xl shadow-lg
                      ${msg.sender === 'user'
                        ? isGreening
                          ? 'bg-emerald-600 text-white rounded-tr-lg'
                          : 'bg-[#e11d48] text-white rounded-tr-lg'
                        : isGreening
                          ? 'bg-emerald-950/40 backdrop-blur-xl border border-emerald-500/15 text-emerald-50 rounded-tl-lg'
                          : 'bg-white/40 dark:bg-white/5 backdrop-blur-xl border border-black/5 dark:border-white/10 text-zinc-900 dark:text-zinc-100 rounded-tl-lg'}
                    `}>
                      {msg.sender === 'bot' && (
                        <span className="absolute -top-1 -left-1">
                          <SentimentDot sentiment={msg.sentiment} />
                        </span>
                      )}
                      {msg.sender === 'bot' && (
                        <span className="pointer-events-none absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-white/15 to-transparent rounded-t-3xl" />
                      )}
                      <p className="text-sm leading-relaxed">{msg.text}</p>
                    </div>
                    <p className={`text-[10px] mt-1.5 font-mono opacity-40 ${msg.sender === 'user' ? 'text-right' : 'text-left'}`}>
                      {msg.timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                    </p>
                  </div>
                </motion.div>
              ))}
            </AnimatePresence>

            {/* Typing indicator */}
            <AnimatePresence>
              {loading && (
                <motion.div
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: 10 }}
                  className="flex gap-3 max-w-4xl mr-auto"
                >
                  <div className={`w-9 h-9 rounded-2xl flex-shrink-0 flex items-center justify-center shadow-lg ${accentBg}`}>
                    <Leaf className="w-4 h-4 text-white" />
                  </div>
                  <div className={`
                    px-5 py-4 rounded-3xl rounded-tl-lg shadow-lg
                    ${isGreening
                      ? 'bg-emerald-950/40 backdrop-blur-xl border border-emerald-500/15 text-emerald-300'
                      : 'bg-white/40 dark:bg-white/5 backdrop-blur-xl border border-black/5 dark:border-white/10 text-zinc-400'}
                  `}>
                    <TypingDots />
                  </div>
                </motion.div>
              )}
            </AnimatePresence>

            <div ref={messagesEndRef} />
          </div>

          {/* Input area */}
          <div className={`px-4 sm:px-8 pb-6 pt-4 border-t ${isGreening ? 'border-emerald-500/10' : 'border-black/5 dark:border-white/5'}`}>
            {/* Crisis disclaimer */}
            <p className={`text-[10px] font-mono text-center mb-4 opacity-50 ${accentClass}`}>
              💚 VANA supports, not replaces, professional care. For immediate help, call 988 (US) or iCall: 9152987821 (India).
            </p>

            <div className={`
              flex items-end gap-3 p-3 rounded-3xl shadow-xl transition-all duration-300
              ${isGreening
                ? 'bg-emerald-950/30 backdrop-blur-xl border border-emerald-500/15'
                : 'bg-white/60 dark:bg-white/5 backdrop-blur-xl border border-black/5 dark:border-white/10'}
            `}>
              {/* Voice button */}
              <motion.button
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                onClick={handleVoice}
                className={`
                  flex-shrink-0 w-11 h-11 rounded-2xl flex items-center justify-center transition-all
                  ${isListening
                    ? `${accentBg} text-white shadow-lg animate-pulse`
                    : 'dark:bg-white/5 bg-black/5 text-zinc-400 hover:text-zinc-200'}
                `}
                title="Voice input"
              >
                {isListening ? <MicOff className="w-5 h-5" /> : <Mic className="w-5 h-5" />}
              </motion.button>

              {/* Text input */}
              <input
                ref={inputRef}
                type="text"
                value={input}
                onChange={e => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder={isListening ? 'Listening...' : 'Share what\'s on your mind...'}
                disabled={loading}
                className={`
                  flex-1 bg-transparent text-sm placeholder-zinc-400 dark:placeholder-zinc-600
                  focus:outline-none resize-none py-2
                  ${isGreening ? 'text-emerald-50' : 'text-zinc-900 dark:text-zinc-100'}
                `}
              />

              {/* Send button */}
              <motion.button
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                onClick={() => sendMessage()}
                disabled={loading || !input.trim()}
                className={`
                  flex-shrink-0 w-11 h-11 rounded-2xl flex items-center justify-center transition-all shadow-lg
                  ${input.trim() && !loading
                    ? `${accentBg} text-white`
                    : 'dark:bg-white/5 bg-black/5 text-zinc-400 cursor-not-allowed opacity-50'}
                `}
                title="Send message"
              >
                <Send className="w-5 h-5" />
              </motion.button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
