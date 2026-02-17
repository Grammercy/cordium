import sys

with open('internal/handlers/channels.go', 'r') as f:
    content = f.read()

# Replacement 1: ChannelListHandler template
search1 = '''	tmpl, err := template.New("channels.html").Funcs(funcMap).ParseFiles(filepath.Join("web", "templates", "channels.html"))
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	data := struct {
		GuildID   string
		GuildName string
		Channels  []*discordgo.Channel
	}{
		GuildID:   guildID,
		GuildName: guildName,
		Channels:  channels,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Printf("Template execution error (channels): %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	w.Write(buf.Bytes())'''

replace1 = '''	data := struct {
		GuildID   string
		GuildName string
		Channels  []*discordgo.Channel
	}{
		GuildID:   guildID,
		GuildName: guildName,
		Channels:  channels,
	}

	if ChannelsTmpl == nil {
		http.Error(w, "Templates not initialized", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := ChannelsTmpl.Execute(&buf, data); err != nil {
		log.Printf("Template execution error (channels): %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	w.Write(buf.Bytes())'''

content = content.replace(search1, replace1)

# Replacement 2: GetMessageHandler template
search2 = '''	tmpl, err := template.New("message.html").Funcs(funcMap).ParseFiles(filepath.Join("web", "templates", "message.html"))
	if err != nil {
		log.Printf("Template parse error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, msg); err != nil {
		log.Printf("Template execute error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	w.Write(buf.Bytes())'''

replace2 = '''	if MessageTmpl == nil {
		http.Error(w, "Templates not initialized", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := MessageTmpl.Execute(&buf, msg); err != nil {
		log.Printf("Template execute error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	w.Write(buf.Bytes())'''

content = content.replace(search2, replace2)

# Replacement 3: ChatViewHandler template
search3 = '''	tmpl, err := template.New("chat.html").Funcs(funcMap).ParseFiles(
		filepath.Join("web", "templates", "chat.html"),
		filepath.Join("web", "templates", "message.html"),
	)
	if err != nil {
		http.Error(w, "Template error: " + err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		GuildID     string
		ChannelID   string
		ChannelName string
		Messages    []*discordgo.Message
		Members     []*discordgo.Member
	}{
		GuildID:     guildID,
		ChannelID:   channelID,
		ChannelName: channelName,
		Messages:    messages,
		Members:     members,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Printf("Template execution error (chat): %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	w.Write(buf.Bytes())'''

replace3 = '''	data := struct {
		GuildID     string
		ChannelID   string
		ChannelName string
		Messages    []*discordgo.Message
		Members     []*discordgo.Member
	}{
		GuildID:     guildID,
		ChannelID:   channelID,
		ChannelName: channelName,
		Messages:    messages,
		Members:     members,
	}

	if ChatTmpl == nil {
		http.Error(w, "Templates not initialized", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := ChatTmpl.Execute(&buf, data); err != nil {
		log.Printf("Template execution error (chat): %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	w.Write(buf.Bytes())'''

content = content.replace(search3, replace3)

with open('internal/handlers/channels.go', 'w') as f:
    f.write(content)
