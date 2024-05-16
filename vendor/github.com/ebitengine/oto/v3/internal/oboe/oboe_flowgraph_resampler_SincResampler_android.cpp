/*
 * Copyright 2019 The Android Open Source Project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#include <cassert>
#include <math.h>
#include "oboe_flowgraph_resampler_SincResampler_android.h"

using namespace RESAMPLER_OUTER_NAMESPACE::resampler;

SincResampler::SincResampler(const MultiChannelResampler::Builder &builder)
        : MultiChannelResampler(builder)
        , mSingleFrame2(builder.getChannelCount()) {
    assert((getNumTaps() % 4) == 0); // Required for loop unrolling.
    mNumRows = kMaxCoefficients / getNumTaps(); // includes guard row
    int32_t numRowsNoGuard = mNumRows - 1;
    mPhaseScaler = (double) numRowsNoGuard / mDenominator;
    double phaseIncrement = 1.0 / numRowsNoGuard;
    generateCoefficients(builder.getInputRate(),
                         builder.getOutputRate(),
                         mNumRows,
                         phaseIncrement,
                         builder.getNormalizedCutoff());
}

void SincResampler::readFrame(float *frame) {
    // Clear accumulator for mixing.
    std::fill(mSingleFrame.begin(), mSingleFrame.end(), 0.0);
    std::fill(mSingleFrame2.begin(), mSingleFrame2.end(), 0.0);

    // Determine indices into coefficients table.
    double tablePhase = getIntegerPhase() * mPhaseScaler;
    int indexLow = static_cast<int>(floor(tablePhase));
    int indexHigh = indexLow + 1; // OK because using a guard row.
    assert (indexHigh < mNumRows);
    float *coefficientsLow = &mCoefficients[static_cast<size_t>(indexLow)
                                            * static_cast<size_t>(getNumTaps())];
    float *coefficientsHigh = &mCoefficients[static_cast<size_t>(indexHigh)
                                             * static_cast<size_t>(getNumTaps())];

    float *xFrame = &mX[static_cast<size_t>(mCursor) * static_cast<size_t>(getChannelCount())];
    for (int tap = 0; tap < mNumTaps; tap++) {
        float coefficientLow = *coefficientsLow++;
        float coefficientHigh = *coefficientsHigh++;
        for (int channel = 0; channel < getChannelCount(); channel++) {
            float sample = *xFrame++;
            mSingleFrame[channel] += sample * coefficientLow;
            mSingleFrame2[channel] += sample * coefficientHigh;
        }
    }

    // Interpolate and copy to output.
    float fraction = tablePhase - indexLow;
    for (int channel = 0; channel < getChannelCount(); channel++) {
        float low = mSingleFrame[channel];
        float high = mSingleFrame2[channel];
        frame[channel] = low + (fraction * (high - low));
    }
}
