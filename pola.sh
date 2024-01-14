#! /bin/bash

f () {
  sleep 8
  echo "zola: after"
}

f &
echo "zola: before"
