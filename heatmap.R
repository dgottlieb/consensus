library(ggplot2)
library(reshape2)
lag <- as.matrix(read.csv("./lag.csv"))
p <- ggplot(melt(lag), aes(Var1, Var2, fill=value)) +
                        geom_tile() +
                        geom_text(aes(label=paste(value)), show_guide = FALSE) +
                        scale_fill_gradient2(midpoint=0,
                                             low="#B2182B",
                                             high="#2166AC",
                                             name="lag (seconds)") +
                        xlab("") + ylab("")
ggsave(filename="lag.png", p, path="./img")
