#!/bin/bash
# Smart worker script that runs appropriate mode based on day/time
# Optimized for game days with live updates

set -e

PIPELINE="target/release/nfl-data-pipeline"

# Get current day of week (0=Sunday, 1=Monday, etc.)
DAY=$(date +%u)
# Get current hour (0-23)
HOUR=$(date +%H)

echo "🏈 NFL Data Pipeline Worker Starting"
echo "📅 Day: $DAY, Hour: $HOUR"

# Continuous loop for always-on worker
while true; do
    # Sunday (DAY=7): Game day - run live mode during game hours
    if [ "$DAY" -eq 7 ]; then
        if [ "$HOUR" -ge 13 ] && [ "$HOUR" -le 23 ]; then
            echo "🔴 SUNDAY GAME DAY - Running live mode"
            $PIPELINE --mode live || echo "⚠️  Live mode failed, continuing..."
        else
            echo "📊 Sunday off-hours - Running update mode"
            $PIPELINE --mode update || echo "⚠️  Update failed, continuing..."
        fi

    # Monday (DAY=1): Monday Night Football
    elif [ "$DAY" -eq 1 ]; then
        if [ "$HOUR" -ge 20 ] && [ "$HOUR" -le 23 ]; then
            echo "🏈 MONDAY NIGHT FOOTBALL - Running live mode"
            $PIPELINE --mode live || echo "⚠️  Live mode failed, continuing..."
        else
            echo "📊 Monday off-hours - Running update mode"
            $PIPELINE --mode update || echo "⚠️  Update failed, continuing..."
        fi

    # Thursday (DAY=4): Thursday Night Football
    elif [ "$DAY" -eq 4 ]; then
        if [ "$HOUR" -ge 20 ] && [ "$HOUR" -le 23 ]; then
            echo "🏈 THURSDAY NIGHT FOOTBALL - Running live mode"
            $PIPELINE --mode live || echo "⚠️  Live mode failed, continuing..."
        else
            echo "📊 Thursday off-hours - Running update mode"
            $PIPELINE --mode update || echo "⚠️  Update failed, continuing..."
        fi

    # All other days/times: Regular update mode
    else
        echo "📊 Regular day - Running update mode"
        $PIPELINE --mode update || echo "⚠️  Update failed, continuing..."
    fi

    # Sleep between runs
    # During game hours: 5 minutes (frequent updates)
    # Off hours: 30 minutes (less frequent)
    if [ "$DAY" -eq 7 ] && [ "$HOUR" -ge 13 ] && [ "$HOUR" -le 23 ]; then
        # Sunday game day (1 PM - 11 PM)
        echo "⏱️  Sleeping 5 minutes (game day)..."
        sleep 300  # 5 minutes
    elif ([ "$DAY" -eq 1 ] || [ "$DAY" -eq 4 ]) && [ "$HOUR" -ge 20 ] && [ "$HOUR" -le 23 ]; then
        # Monday/Thursday night football (8 PM - 11 PM)
        echo "⏱️  Sleeping 5 minutes (game day)..."
        sleep 300  # 5 minutes
    else
        echo "⏱️  Sleeping 30 minutes (off hours)..."
        sleep 1800  # 30 minutes
    fi

    # Refresh day/hour for next iteration
    DAY=$(date +%u)
    HOUR=$(date +%H)
done
