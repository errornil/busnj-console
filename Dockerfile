FROM scratch

EXPOSE 6004

VOLUME /receipts

ADD transactions.html /transactions.html
ADD history.html /history.html
ADD receipt.html /receipt.html

ADD main.bin /main

CMD ["/main"]
