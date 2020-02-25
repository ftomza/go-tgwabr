package tg

import (
	"context"
	"testing"
	"tgwabr/api"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func TestService_prepareArgs(t *testing.T) {
	type fields struct {
		ctx        context.Context
		bot        *tgbotapi.BotAPI
		mainGroups []int64
		TG         api.TG
	}
	type args struct {
		args string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantArg1 string
		wantArg2 string
	}{
		{name: "Phone MG", fields: fields{}, args: args{"+7(911) 113 59 00 dubai"}, wantArg1: "79111135900", wantArg2: "dubai"},
		{name: "Phone2 MG", fields: fields{}, args: args{"+7(911)113-59-00 dubai"}, wantArg1: "79111135900", wantArg2: "dubai"},
		{name: "Phone3 MG", fields: fields{}, args: args{"+79111135900 dubai"}, wantArg1: "79111135900", wantArg2: "dubai"},
		{name: "Phone4 MG", fields: fields{}, args: args{"79111135900 dubai"}, wantArg1: "79111135900", wantArg2: "dubai"},
		{name: "Phone", fields: fields{}, args: args{"+7(911) 113 59 00"}, wantArg1: "79111135900", wantArg2: ""},
		{name: "Phone2", fields: fields{}, args: args{"+7(911)113-59-00"}, wantArg1: "79111135900", wantArg2: ""},
		{name: "Phone3", fields: fields{}, args: args{"+79111135900"}, wantArg1: "79111135900", wantArg2: ""},
		{name: "Phone4", fields: fields{}, args: args{"79111135900"}, wantArg1: "79111135900", wantArg2: ""},
		{name: "Alias MG", fields: fields{}, args: args{"artem dubai"}, wantArg1: "artem", wantArg2: "dubai"},
		{name: "Alias", fields: fields{}, args: args{"artem"}, wantArg1: "artem", wantArg2: ""},
		{name: "Group", fields: fields{}, args: args{"13258-3698"}, wantArg1: "13258-3698", wantArg2: ""},
		{name: "Group MG", fields: fields{}, args: args{"13258-3698 dubai"}, wantArg1: "13258-3698", wantArg2: "dubai"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				ctx:        tt.fields.ctx,
				bot:        tt.fields.bot,
				mainGroups: tt.fields.mainGroups,
				TG:         tt.fields.TG,
			}
			gotArg1, gotArg2 := s.prepareArgs(tt.args.args)
			gotArg1 = s.prepareClient(gotArg1)
			if gotArg1 != tt.wantArg1 {
				t.Errorf("prepareArgs() gotArg1 = %v, want %v", gotArg1, tt.wantArg1)
			}
			if gotArg2 != tt.wantArg2 {
				t.Errorf("prepareArgs() gotArg2 = %v, want %v", gotArg2, tt.wantArg2)
			}
		})
	}
}
