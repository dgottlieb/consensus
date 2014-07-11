library(ggplot2)
library(reshape2)
lag <- as.matrix(read.csv("./lag.csv"))
p <- ggplot(melt(lag)) + geom_tile(aes(Var1, Var2, fill=value)) +
                         scale_fill_gradient(low = "white", high = "steelblue", name="lag (seconds)") +
                         xlab("") + ylab("")
ggsave(filename="lag.png", p, path="./img")
