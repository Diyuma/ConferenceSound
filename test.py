import wave
import numpy as np

wav_obj = wave.open('output.wav', 'rb')

sample_freq = wav_obj.getframerate()
n_samples = wav_obj.getnframes()

t_audio = n_samples/sample_freq

signal_wave = wav_obj.readframes(n_samples)
signal_array = np.frombuffer(signal_wave, dtype=np.int16)

l_channel = signal_array[0::2]
r_channel = signal_array[1::2]
print(l_channel)
print(r_channel)
print(t_audio)
print(np.min(r_channel))

#times = np.linspace(0, n_samples/sample_freq, num=n_samples)

import matplotlib.pyplot as plt
plt.figure(figsize=(15, 5))
plt.plot([i / 10000 for i in range(len(l_channel))], l_channel)
plt.title('Left Channel')
plt.ylabel('Signal Value')
plt.xlabel('Time (s)')
plt.xlim(0, t_audio)
plt.show()