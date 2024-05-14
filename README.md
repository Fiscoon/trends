# SRE Trends calculation backend

The goal of this program is to get the current status of different clusters in a human-readable manner.

## How it works

1. Queries Prometheus to get all the hosts from the clusters defined at trends.go.
2. Gets ALL the data points for the hosts' CPU usage from Prometheus. (One data point per minute)
3. Goes through each data point to see if they cross over 70%, 90% and 99.5% CPU usage.
4. Assigns a score to each cluster based on how much did its hosts cross 70%, 90% and 99.5% CPU usage.
5. Calculates the current state of the cluster based on the score. The status can be good, average or bad.

## Why use scoring to calculate the status of a cluster?

Basically, the more points a cluster gets, the worse is its state.

This system has some advantages, we can have a very flexible criteria to define when a cluster is good and when it's bad. 

For example, if we believe that having a host constantly over 70% CPU is something very bad, we can just raise the number of points that a cluster gets when a host is over 70%. Or if we believe that a host reaching 70% CPU is not a big deal, then we can just lower the amount of points that gets added when a host goes over 70%.

**Current scoring definition:**
* **Green - Good:** If the cluster has less than 10 points
* **Yellow - Average:** If the cluster has less than 40 points
* **Red - Bad:** If the cluster has 40 points or more

## Usage

go build -o trendsbin && ./trendsbin
Then, query localhost:8080/summary or localhost:8080/trends
