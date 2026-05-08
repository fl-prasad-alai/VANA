import React, { useEffect, useRef } from 'react';
import * as THREE from 'three';
import { useTheme } from '../contexts/ThemeContext';

const THEME_COLORS = {
  greening: {
    base: new THREE.Color(0.043, 0.102, 0.055), // #0B1A0E
    mid: new THREE.Color(0.176, 0.353, 0.153),  // #2D5A27
    highlight: new THREE.Color(0.290, 0.545, 0.259), // #4A8B42
    bgClass: "bg-[#0B1A0E]"
  },
  dark: {
    base: new THREE.Color(0.02, 0.02, 0.03), 
    mid: new THREE.Color(0.1, 0.06, 0.23), 
    highlight: new THREE.Color(0.3, 0.16, 0.52), 
    bgClass: "bg-[#050508]"
  },
  light: {
    base: new THREE.Color(0.97, 0.98, 0.99), 
    mid: new THREE.Color(0.88, 0.91, 0.94), 
    highlight: new THREE.Color(0.73, 0.9, 0.99), 
    bgClass: "bg-slate-50"
  }
};

const LiquidForest: React.FC = () => {
  const { theme } = useTheme();
  const containerRef = useRef<HTMLDivElement>(null);
  const rendererRef = useRef<THREE.WebGLRenderer | null>(null);
  const materialRef = useRef<THREE.ShaderMaterial | null>(null);
  const animationFrameRef = useRef<number>();

  useEffect(() => {
    if (!containerRef.current) return;

    // --- Scene Setup ---
    const scene = new THREE.Scene();
    const camera = new THREE.OrthographicCamera(-1, 1, 1, -1, 0, 1);
    
    const renderer = new THREE.WebGLRenderer({ 
      antialias: true, 
      alpha: true,
      powerPreference: "high-performance" 
    });
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2)); // Cap pixel ratio for performance
    renderer.setSize(window.innerWidth, window.innerHeight);
    containerRef.current.appendChild(renderer.domElement);
    rendererRef.current = renderer;

    const geometry = new THREE.PlaneGeometry(2, 2);

    // --- High-Fidelity Shader Engineering (GLSL) ---
    const vertexShader = `
      varying vec2 vUv;
      void main() {
        vUv = uv;
        gl_Position = vec4(position, 1.0);
      }
    `;

    const fragmentShader = `
      uniform float uTime;
      uniform vec2 uResolution;
      uniform vec3 uColorBase;
      uniform vec3 uColorMid;
      uniform vec3 uColorHighlight;
      varying vec2 vUv;

      // Hash function for noise
      vec2 hash(vec2 p) {
        p = vec2(dot(p, vec2(127.1, 311.7)), dot(p, vec2(269.5, 183.3)));
        return -1.0 + 2.0 * fract(sin(p) * 43758.5453123);
      }

      // Simplex Noise
      float noise(vec2 p) {
        const float K1 = 0.366025404; // (sqrt(3)-1)/2;
        const float K2 = 0.211324865; // (3-sqrt(3))/6;

        vec2 i = floor(p + (p.x + p.y) * K1);
        vec2 a = p - i + (i.x + i.y) * K2;
        vec2 o = (a.x > a.y) ? vec2(1.0, 0.0) : vec2(0.0, 1.0);
        vec2 b = a - o + K2;
        vec2 c = a - 1.0 + 2.0 * K2;

        vec3 h = max(0.5 - vec3(dot(a,a), dot(b,b), dot(c,c)), 0.0);

        vec3 n = h * h * h * h * vec3(dot(a, hash(i + 0.0)), dot(b, hash(i + o)), dot(c, hash(i + 1.0)));
        return dot(n, vec3(70.0));
      }

      // Fractional Brownian Motion for layered water ripples
      float fbm(vec2 p) {
        float f = 0.0;
        float w = 0.5;
        for (int i = 0; i < 5; i++) {
          f += w * noise(p);
          p *= 2.0;
          w *= 0.5;
        }
        return f;
      }

      void main() {
        vec2 uv = vUv;
        // Correct aspect ratio
        float ratio = uResolution.x / uResolution.y;
        vec2 p = uv * 2.0 - 1.0;
        p.x *= ratio;

        // Extremely slow meditative time scale
        float t = uTime * 0.15;

        // Create flowing water distortions
        vec2 flow = vec2(
          fbm(p * 1.5 + vec2(t * 0.2, t * 0.1)),
          fbm(p * 1.5 - vec2(t * 0.1, t * 0.3))
        );

        // Calculate caustic patterns using distorted coordinates
        float caustics = fbm(p * 3.0 + flow * 2.0 + t * 0.4);
        
        // Sharpen the peaks to create the "light reflection" water ripple effect
        caustics = abs(caustics);
        caustics = 1.0 - caustics;
        caustics = pow(caustics, 3.0); // Intensity of the ripples

        // Background ambient movement
        float ambient = fbm(p * 1.0 - t * 0.1);

        // Mix colors based on the ripples and ambient depth
        vec3 finalColor = uColorBase;
        finalColor = mix(finalColor, uColorMid, ambient * 0.5 + 0.5);
        finalColor = mix(finalColor, uColorHighlight, caustics * 0.8);

        // Vignette effect to darken edges
        float vignette = length(p);
        vignette = smoothstep(1.5, 0.5, vignette);
        finalColor *= vignette;

        gl_FragColor = vec4(finalColor, 1.0);
      }
    `;

    const initialColors = THEME_COLORS[theme] || THEME_COLORS.greening;

    const material = new THREE.ShaderMaterial({
      vertexShader,
      fragmentShader,
      uniforms: {
        uTime: { value: 0 },
        uResolution: { value: new THREE.Vector2(window.innerWidth, window.innerHeight) },
        uColorBase: { value: initialColors.base },
        uColorMid: { value: initialColors.mid },
        uColorHighlight: { value: initialColors.highlight }
      },
      transparent: true,
      depthWrite: false,
      depthTest: false
    });
    materialRef.current = material;

    const mesh = new THREE.Mesh(geometry, material);
    scene.add(mesh);

    // --- Performance Optimization ---
    let isVisible = true;
    const handleVisibilityChange = () => {
      isVisible = !document.hidden;
    };
    document.addEventListener("visibilitychange", handleVisibilityChange);

    const handleResize = () => {
      const width = window.innerWidth;
      const height = window.innerHeight;
      renderer.setSize(width, height);
      material.uniforms.uResolution.value.set(width, height);
    };
    window.addEventListener('resize', handleResize);

    const animate = (time: number) => {
      if (isVisible) {
        material.uniforms.uTime.value = time * 0.001;
        renderer.render(scene, camera);
      }
      animationFrameRef.current = requestAnimationFrame(animate);
    };
    animationFrameRef.current = requestAnimationFrame(animate);

    return () => {
      window.removeEventListener('resize', handleResize);
      document.removeEventListener("visibilitychange", handleVisibilityChange);
      if (animationFrameRef.current) cancelAnimationFrame(animationFrameRef.current);
      renderer.dispose();
      geometry.dispose();
      material.dispose();
      if (containerRef.current) {
        containerRef.current.removeChild(renderer.domElement);
      }
    };
  }, []); // Run once on mount

  // Update uniforms when theme changes
  useEffect(() => {
    if (materialRef.current) {
      const colors = THEME_COLORS[theme] || THEME_COLORS.greening;
      materialRef.current.uniforms.uColorBase.value.copy(colors.base);
      materialRef.current.uniforms.uColorMid.value.copy(colors.mid);
      materialRef.current.uniforms.uColorHighlight.value.copy(colors.highlight);
    }
  }, [theme]);

  const bgClass = THEME_COLORS[theme]?.bgClass || THEME_COLORS.greening.bgClass;

  return (
    <div 
      ref={containerRef} 
      className={`fixed inset-0 -z-10 ${bgClass} overflow-hidden transition-colors duration-1000`}
      style={{ filter: 'blur(3px) scale(1.05)' }} // Slight 3D depth blur
    />
  );
};

export default LiquidForest;
