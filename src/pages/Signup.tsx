import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { User, Mail, Lock, ArrowRight, Leaf, ShieldCheck } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';
import { useTheme } from '../contexts/ThemeContext';
import LiquidForest from '../components/LiquidForest';
import ThemeToggle from '../components/ThemeToggle';

const FADE_UP = {
  hidden: { opacity: 0, y: 16 },
  show: { opacity: 1, y: 0, transition: { type: 'spring' as const, stiffness: 300, damping: 28 } },
};
const STAGGER = { show: { transition: { staggerChildren: 0.06, delayChildren: 0.1 } } };

function Field({
  label, icon: Icon, type, value, onChange, autoComplete, placeholder, error, isGreening,
}: {
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  type: string;
  value: string;
  onChange: (v: string) => void;
  autoComplete?: string;
  placeholder: string;
  error?: string;
  isGreening: boolean;
}) {
  return (
    <motion.div variants={FADE_UP} className="text-left">
      <label className={`block text-[10px] font-bold tracking-[0.2em] uppercase ml-1 mb-2 ${isGreening ? 'text-emerald-300' : 'text-zinc-500 dark:text-zinc-400'}`}>
        {label}
      </label>
      <div className="relative">
        <Icon className={`absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 transition-colors ${isGreening ? 'text-emerald-400' : 'text-zinc-400'}`} />
        <input
          type={type}
          value={value}
          onChange={e => onChange(e.target.value)}
          autoComplete={autoComplete}
          className={`
            w-full py-3.5 pl-12 pr-4 rounded-2xl text-sm font-medium transition-all focus:outline-none 
            ${isGreening 
              ? 'bg-emerald-950/40 border border-emerald-500/30 text-white placeholder-emerald-500/50 focus:border-emerald-500/70 focus:bg-emerald-900/60 shadow-inner' 
              : 'bg-white/60 dark:bg-black/40 border border-black/5 dark:border-white/10 text-zinc-900 dark:text-zinc-100 placeholder-zinc-400 dark:placeholder-zinc-600 focus:border-violet-500/50'}
            ${error ? 'border-red-500/50 focus:border-red-500/50' : ''}
          `}
          placeholder={placeholder}
        />
      </div>
      {error && <p className="text-xs text-red-400 mt-1">{error}</p>}
    </motion.div>
  );
}

export const SignupPage: React.FC = () => {
  const { register, isAuthenticated } = useAuth();
  const { theme } = useTheme();
  const navigate = useNavigate();
  const isGreening = theme === 'greening';

  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [formError, setFormError] = useState('');
  const [loading, setLoading] = useState(false);
  const [passwordStrength, setPasswordStrength] = useState(0);

  useEffect(() => {
    if (isAuthenticated) navigate('/dashboard');
  }, [isAuthenticated, navigate]);

  const checkStrength = (pwd: string): number => {
    let s = 0;
    if (pwd.length >= 8) s++;
    if (pwd.length >= 12) s++;
    if (/[a-z]/.test(pwd) && /[A-Z]/.test(pwd)) s++;
    if (/\d/.test(pwd)) s++;
    if (/[^a-zA-Z\d]/.test(pwd)) s++;
    return Math.min(s, 5);
  };

  const handlePasswordChange = (v: string) => {
    setPassword(v);
    setPasswordStrength(checkStrength(v));
  };

  const strengthColor = passwordStrength <= 2 ? '#ef4444' : passwordStrength === 3 ? '#f59e0b' : '#10b981';
  const strengthLabel = ['', 'Weak', 'Weak', 'Good', 'Strong', 'Strong'][passwordStrength];

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFieldErrors({});
    setFormError('');

    const errs: Record<string, string> = {};
    if (!name.trim())           errs.name     = 'Name is required';
    if (!email.trim())          errs.email    = 'Email is required';
    else if (!email.includes('@')) errs.email = 'Enter a valid email';
    if (password.length < 8)   errs.password = 'Minimum 8 characters';

    if (Object.keys(errs).length) { setFieldErrors(errs); return; }

    setLoading(true);
    try {
      const result = await register(email.trim(), password, name.trim());
      if (!result.success) {
        setFormError(result.error || 'Something went wrong. Please try again.');
      }
    } catch {
      setFormError('Something went wrong. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className={`relative min-h-screen flex flex-col transition-colors duration-500 ${isGreening ? 'bg-transparent text-[#ecfdf5]' : 'bg-transparent dark:bg-transparent text-zinc-900 dark:text-zinc-100'}`}>
      <LiquidForest />

      {/* Theme Toggle */}
      <div className="fixed bottom-8 left-1/2 -translate-x-1/2 sm:top-8 sm:right-8 sm:bottom-auto sm:left-auto sm:translate-x-0 z-50 scale-[0.85] sm:scale-100">
        <ThemeToggle />
      </div>

      <div className="flex-1 flex items-center justify-center px-4 py-16 relative z-10">
        <motion.div
          variants={STAGGER}
          initial="hidden"
          animate="show"
          className="w-full max-w-sm"
        >
          {/* Logo */}
          <motion.div variants={FADE_UP} className="flex items-center gap-2 justify-center mb-8">
            <div className={`w-9 h-9 rounded-xl flex items-center justify-center shadow-lg ${isGreening ? 'bg-emerald-600 shadow-emerald-600/20' : 'bg-[#e11d48] shadow-[#e11d48]/20'}`}>
              <Leaf className="w-5 h-5 text-white" />
            </div>
            <h1 className="text-xl font-black tracking-tighter dark:text-zinc-100 text-zinc-900 font-heading">
              VANA <span className={isGreening ? 'text-emerald-500' : 'text-[#e11d48]'}>MIND</span>
            </h1>
          </motion.div>

          {/* Card */}
          <motion.div
            variants={FADE_UP}
            className={`
              relative overflow-hidden rounded-[30px] p-8 sm:p-10 shadow-2xl transition-all duration-500
              ${isGreening
                ? 'bg-emerald-950/20 backdrop-blur-[80px] border border-emerald-500/20'
                : 'bg-white/40 dark:bg-white/5 border border-black/5 dark:border-white/10 backdrop-blur-[60px]'}
            `}
          >
            <div className={`absolute -top-24 -right-24 w-48 h-48 blur-[60px] rounded-full pointer-events-none ${isGreening ? 'bg-emerald-500/15' : 'bg-[#e11d48]/5'}`} />

            <motion.h1 variants={FADE_UP} className="text-2xl font-black tracking-tighter mb-1 font-heading">
              Create account
            </motion.h1>
            <motion.p variants={FADE_UP} className={`text-sm mb-8 italic ${isGreening ? 'text-emerald-100/70' : 'dark:text-zinc-400 text-zinc-500'}`}>
              Begin your mental wellness journey with VANA
            </motion.p>

            <form onSubmit={handleSubmit} noValidate className="space-y-3">
              <Field
                label="Full Name" icon={User} type="text" value={name} onChange={setName}
                autoComplete="name" placeholder="Arjun Mehta" error={fieldErrors.name} isGreening={isGreening}
              />
              <Field
                label="Email" icon={Mail} type="email" value={email} onChange={setEmail}
                autoComplete="email" placeholder="you@example.com" error={fieldErrors.email} isGreening={isGreening}
              />
              <Field
                label="Password" icon={Lock} type="password" value={password} onChange={handlePasswordChange}
                autoComplete="new-password" placeholder="8+ characters" error={fieldErrors.password} isGreening={isGreening}
              />

              {/* Password strength */}
              <AnimatePresence>
                {password && (
                  <motion.div
                    initial={{ opacity: 0, height: 0 }}
                    animate={{ opacity: 1, height: 'auto' }}
                    exit={{ opacity: 0, height: 0 }}
                    className="overflow-hidden"
                  >
                    <div className="flex gap-1 mb-1">
                      {[...Array(5)].map((_, i) => (
                        <div
                          key={i}
                          className="h-1 flex-1 rounded-full transition-colors duration-300"
                          style={{ backgroundColor: i < passwordStrength ? strengthColor : 'rgba(255,255,255,0.1)' }}
                        />
                      ))}
                    </div>
                    <p className="text-xs" style={{ color: strengthColor }}>{strengthLabel}</p>
                  </motion.div>
                )}
              </AnimatePresence>

              {formError && (
                <motion.div
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  className="text-sm text-red-400 bg-red-500/10 border border-red-500/20 rounded-xl px-3 py-2 flex items-center gap-2"
                >
                  <ShieldCheck className="w-4 h-4 flex-shrink-0" />
                  {formError}
                </motion.div>
              )}

              <motion.div variants={FADE_UP} className="pt-1">
                <motion.button
                  type="submit"
                  disabled={loading}
                  whileTap={{ scale: 0.97, transition: { type: 'spring', stiffness: 600, damping: 35 } }}
                  className={`
                    w-full flex items-center justify-center gap-2
                    py-2.5 rounded-xl text-sm font-bold text-white
                    disabled:opacity-50 disabled:cursor-not-allowed
                    transition-all shadow-xl uppercase tracking-widest text-xs
                    ${isGreening ? 'bg-emerald-600 shadow-emerald-600/20' : 'bg-[#e11d48] shadow-[#e11d48]/20'}
                  `}
                >
                  {loading ? (
                    <span className="flex items-center gap-2">
                      <motion.span
                        animate={{ rotate: 360 }}
                        transition={{ duration: 0.8, repeat: Infinity, ease: 'linear' }}
                        className="block w-4 h-4 border-2 border-white/30 border-t-white rounded-full"
                      />
                      Creating…
                    </span>
                  ) : (
                    <>Start Your Journey <ArrowRight className="w-4 h-4" /></>
                  )}
                </motion.button>
              </motion.div>
            </form>
          </motion.div>

          <motion.p variants={FADE_UP} className={`text-center text-sm mt-6 ${isGreening ? 'text-emerald-100/70' : 'dark:text-zinc-400 text-zinc-500'}`}>
            Already have an account?{' '}
            <Link to="/login" className={`font-bold hover:underline transition-colors ${isGreening ? 'text-emerald-400 hover:text-emerald-300' : 'text-[#e11d48]'}`}>
              Sign in
            </Link>
          </motion.p>
        </motion.div>
      </div>
    </div>
  );
};
